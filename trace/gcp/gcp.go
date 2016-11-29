// Package gcp contains a Google Cloud Platform-specific implementation of the generic tracing APIs.
package gcp

import (
	"context"
	"fmt"
	"log"
	"time"

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

type Client struct {
	service *api.Service
	proj    string
	bundler *bundler.Bundler
}

func NewClient(ctx context.Context, projID string, opts ...option.ClientOption) (*Client, error) {
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
	c := &Client{
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

func (c *Client) upload(traces []*api.Trace) error {
	_, err := c.service.Projects.PatchTraces(c.proj, &api.Traces{Traces: traces}).Do()
	return err
}

func (c *Client) NewSpan(ctx context.Context, name string) context.Context {
	parent := contextSpan(ctx)
	s := &span{
		id:   nextSpanID(),
		name: name,
	}
	if parent == nil {
		s.trace = &trace{
			id:     nextTraceID(),
			spans:  make([]*span, 0),
			client: c,
		}
	} else {
		s.trace = parent.trace
		s.parentID = parent.id
	}
	s.start = time.Now()
	s.trace.Lock()
	s.trace.spans = append(s.trace.spans, s)
	s.trace.Unlock()
	return context.WithValue(ctx, spanKey, s)
}

func (c *Client) TraceID(ctx context.Context) string {
	s := contextSpan(ctx)
	if s == nil {
		return "" // TODO(jbd): panic instead? It should never happen.
	}
	return s.trace.id
}

func (c *Client) Finish(ctx context.Context, tags map[string]interface{}) {
	s := contextSpan(ctx)
	if s == nil {
		return
	}
	s.end = time.Now()
	if err := s.trace.finish(ctx, s); err != nil {
		log.Print(err)
	}
}

func (c *Client) Log(ctx context.Context, payload interface{}) error {
	// TODO(jbd): implement
	return nil
}
