package trace

import (
	"context"

	"golang.org/x/net/trace"
)

type dc struct {
	family string
}

// DefaultClient returns a backend that traces and provides a web UI on /debug/requests and /debug/events.
// These paths are registered on the default mux, start a server in order to be able to see them.
func DefaultClient(family string) Client {
	return &dc{family: family}
}

func (d *dc) NewSpan(ctx context.Context, name string) context.Context {
	tr := trace.New(d.family, name)
	return context.WithValue(ctx, defaultTraceKey, tr)
}

func (d *dc) TraceID(ctx context.Context) []byte {
	return []byte("") // not provided by the default client
}

func (d *dc) Finish(ctx context.Context, labels map[string]interface{}) error {
	v := ctx.Value(defaultTraceKey)
	if v == nil {
		return nil
	}
	tr := v.(trace.Trace)
	tr.Finish()
	return nil
}

var defaultTraceKey = contextKey("defaultTrace")

// func (d *dc) Log(ctx context.Context, payload fmt.Stringer) error {
// 	v := ctx.Value(traceKey)
// 	if v == nil {
// 		return nil
// 	}
// 	tr := v.(trace.Trace)
// 	tr.LazyLog(payload, false)
// 	return nil
// }
