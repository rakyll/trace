// Package gcp contains a Google Cloud Platform-specific implementation of the generic tracing APIs.
package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rakyll/trace"

	api "google.golang.org/api/cloudtrace/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/support/bundler"
	"google.golang.org/api/transport"
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

type client struct {
	service *api.Service
	proj    string
	bundler *bundler.Bundler
}

func NewClient(ctx context.Context, projID string, opts ...option.ClientOption) (trace.Client, error) {
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
	c := &client{
		service: apiService,
		proj:    projID,
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
	return c, nil
}

func (c *client) upload(traces []*api.Trace) error {
	_, err := c.service.Projects.PatchTraces(c.proj, &api.Traces{Traces: traces}).Do()
	// TODO(jbd): How to handle errors?
	return err
}

func (c *client) NewSpan(parent context.Context, casual []byte, name string) context.Context {
	parentSpan := contextSpan(parent)
	s := &span{
		id:   nextSpanID(),
		name: name,
	}
	if parentSpan == nil {
		s.trace = &gcpTrace{
			id:     nextTraceID(),
			spans:  make([]*span, 0),
			client: c,
		}
	} else {
		s.trace = parentSpan.trace
		s.parentID = parentSpan.id
	}
	s.trace.Lock()
	s.trace.spans = append(s.trace.spans, s)
	s.trace.Unlock()

	s.start = time.Now()
	return context.WithValue(parent, spanKey, s)
}

func (c *client) Span(ctx context.Context, name string, info []byte) (context.Context, error) {
	v := spanID{}
	if err := json.Unmarshal(info, &v); err != nil {
		return nil, err
	}
	s := &span{
		trace: &gcpTrace{
			// Resurrected span cannot be finished,
			// hence we don't care about setting all the fields.
			client: c,
			id:     v.TraceID,
		},
		parentID: v.ParentID,
		id:       v.ID,
		name:     name,
	}
	return context.WithValue(ctx, spanKey, s), nil
}

func (c *client) Info(ctx context.Context) []byte {
	s := contextSpan(ctx)
	if s == nil {
		return nil
	}
	return []byte(s.trace.id)
}

func (c *client) Finish(ctx context.Context, tags map[string][]byte) error {
	s := contextSpan(ctx)
	if s == nil {
		return nil
	}
	s.end = time.Now()
	return s.trace.finish(ctx, s)
}
