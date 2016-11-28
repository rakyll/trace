package trace

import "context"
import "fmt"

type Tracer interface {
	NewSpan(ctx context.Context, name string) context.Context
	Finish(ctx context.Context, tags map[string]interface{})
	Log(ctx context.Context, payload interface{}) error
}

func WithTrace(ctx context.Context, t Tracer) context.Context {
	return context.WithValue(ctx, traceKey, t)
}

func WithSpan(ctx context.Context, name string) context.Context {
	t := tracerFromContext(ctx)
	if t == nil {
		return ctx
	}
	return t.NewSpan(ctx, name)
}

func Logf(ctx context.Context, format string, arg ...interface{}) error {
	t := tracerFromContext(ctx)
	return t.Log(ctx, fmt.Sprintf(format, arg))
}

func Finish(ctx context.Context) {
	t := tracerFromContext(ctx)
	if t == nil {
		return
	}
	t.Finish(ctx, nil)
	// TODO(jbd): add tags.
}

func tracerFromContext(ctx context.Context) Tracer {
	v := ctx.Value(traceKey)
	if v == nil {
		return nil
	}
	return v.(Tracer)
}

type contextKey string

var (
	traceKey = contextKey("trace")
)
