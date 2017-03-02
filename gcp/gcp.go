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
	return err
}

func (c *client) NewSpan(parent []byte, causal []byte) []byte { // TODO(jbd): add error.
	var parentID spanID
	var causalID spanID

	if parent != nil {
		json.Unmarshal(parent, &parentID) // ignore errors
	}
	if causal != nil {
		json.Unmarshal(causal, &causalID) // ignore errors
	}

	id := spanID{
		TraceID:  parentID.TraceID,
		ID:       nextSpanID(),
		CausalID: causalID.ID,
		ParentID: parentID.ID,
	}
	by, _ := json.Marshal(id)
	return by
}

func (c *client) Finish(id []byte, name string, labels map[string][]byte, start, end time.Time) error {
	var ident spanID
	if err := json.Unmarshal(id, &ident); err != nil {
		return err
	}
	s := &span{
		name:   name,
		id:     ident,
		labels: labels,
		start:  start,
		end:    end,
	}
	return finish(c, c.proj, []*span{s})
}
