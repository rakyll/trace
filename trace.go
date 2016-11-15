package trace

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"cloud.google.com/go/logging"

	api "google.golang.org/api/cloudtrace/v1"
	"google.golang.org/api/support/bundler"
	"google.golang.org/api/transport"

	"google.golang.org/api/option"
)

const (
	httpHeader          = `X-Cloud-Trace-Context`
	userAgent           = `gcloud-golang-trace/20160501`
	cloudPlatformScope  = `https://www.googleapis.com/auth/cloud-platform`
	spanKindClient      = `RPC_CLIENT`
	spanKindServer      = `RPC_SERVER`
	spanKindUnspecified = `SPAN_KIND_UNSPECIFIED`
	maxStackFrames      = 20
	labelSamplingPolicy = `trace.cloud.google.com/sampling_policy`
	labelSamplingWeight = `trace.cloud.google.com/sampling_weight`
)

type Client struct {
	service   *api.Service
	logClient *logging.Client

	proj    string
	bundler *bundler.Bundler
}

func New(ctx context.Context, projID string, opts ...option.ClientOption) (context.Context, error) {
	o := []option.ClientOption{
		option.WithScopes(cloudPlatformScope),
		option.WithUserAgent(userAgent),
	}
	o = append(o, opts...)
	hc, basePath, err := transport.NewHTTPClient(ctx, o...)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client for Google Stackdriver Trace API: %v", err)
	}
	apiService, err := api.New(hc)
	if err != nil {
		return nil, fmt.Errorf("creating Google Stackdriver Trace API client: %v", err)
	}
	if basePath != "" {
		// An option set a basepath, so override api.New's default.
		apiService.BasePath = basePath
	}
	lc, err := logging.NewClient(ctx, projID)
	if err != nil {
		return nil, err
	}
	c := &Client{
		service:   apiService,
		proj:      projID,
		logClient: lc,
	}
	bundler := bundler.NewBundler((*api.Trace)(nil), func(bundle interface{}) {
		traces := bundle.([]*api.Trace)
		err := c.upload(traces)
		if err != nil {
			log.Printf("failed to upload %d traces to the Cloud Trace server.", len(traces))
		}
	})
	bundler.DelayThreshold = 2 * time.Second
	bundler.BundleCountThreshold = 100
	// We're not measuring bytes here, we're counting traces and spans as one "byte" each.
	bundler.BundleByteThreshold = 1000
	bundler.BundleByteLimit = 1000
	bundler.BufferedByteLimit = 10000
	c.bundler = bundler
	return context.WithValue(ctx, clientKey, c), nil
}

func (c *Client) upload(traces []*api.Trace) error {
	_, err := c.service.Projects.PatchTraces(c.proj, &api.Traces{Traces: traces}).Do()
	return err
}

type trace struct {
	id string

	sync.Mutex
	spans []*Span
}

func (t *trace) finish(ctx context.Context, s *Span) error {
	s.end = time.Now()
	c := contextClient(ctx)
	return c.upload([]*api.Trace{t.constructTrace(c.proj, t.spans)})
}

func (t *trace) constructTrace(projID string, spans []*Span) *api.Trace {
	apiSpans := make([]*api.TraceSpan, len(spans))
	for i, sp := range spans {
		s := &api.TraceSpan{
			Name:         sp.name,
			SpanId:       sp.id,
			ParentSpanId: sp.parentID,
			StartTime:    sp.start.In(time.UTC).Format(time.RFC3339Nano),
			EndTime:      sp.end.In(time.UTC).Format(time.RFC3339Nano),
			// TODO(jbd): add labels
		}
		apiSpans[i] = s
	}
	return &api.Trace{
		ProjectId: projID,
		TraceId:   t.id,
		Spans:     apiSpans,
	}
}

type Span struct {
	client *Client
	trace  *trace

	parentID uint64
	id       uint64
	name     string

	start time.Time
	end   time.Time
}

func WithSpan(ctx context.Context, name string) context.Context {
	parent := ContextSpan(ctx)
	s := &Span{
		id:    nextSpanID(),
		name:  name,
		start: time.Now(),
	}
	if parent == nil {
		s.trace = &trace{
			id:    nextTraceID(),
			spans: make([]*Span, 0),
		}
	} else {
		s.trace = parent.trace
		s.parentID = parent.id
	}
	s.trace.Lock()
	s.trace.spans = append(s.trace.spans, s)
	s.trace.Unlock()

	if s.name == "" {
		// the name of the caller function
		pc, _, _, ok := runtime.Caller(1)
		if ok {
			fn := runtime.FuncForPC(pc)
			s.name = fn.Name()
		}
	}
	return context.WithValue(ctx, ctxKey, s)
}

func Logf(ctx context.Context, format string, arg ...interface{}) error {
	c := contextClient(ctx)
	if c == nil {
		panic("no trace client in context")
	}
	return c.logClient.Logger("").LogSync(ctx, logging.Entry{Payload: fmt.Sprintf(format, arg)})
}

// TraceID returns the ID of the trace to which s belongs.
func (s *Span) TraceID() string {
	if s == nil {
		return ""
	}
	return s.trace.id
}

// ContextSpan returns a span from the current context or nil
// if no contexts exists in the current context.
func ContextSpan(ctx context.Context) *Span {
	v := ctx.Value(ctxKey)
	if v == nil {
		return nil
	}
	return v.(*Span)
}

func contextClient(ctx context.Context) *Client {
	v := ctx.Value(clientKey)
	if v == nil {
		return nil
	}
	return v.(*Client)
}

// Finish finishes the current span.
func Finish(ctx context.Context) {
	s := ContextSpan(ctx)
	if s == nil {
		return
	}
	s.end = time.Now()
	// TODO(jbd):
	// TODO(jbd): Add an error handler?
	if err := s.trace.finish(ctx, s); err != nil {
		log.Print(err)
	}
}

type key string

var ctxKey = key("span")
var clientKey = key("client")
