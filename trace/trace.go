package trace

import (
	"context"
	"fmt"
	"runtime"
)

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
	if name == "" {
		// the name of the caller function
		pc, _, _, ok := runtime.Caller(1)
		if ok {
			fn := runtime.FuncForPC(pc)
			name = fn.Name()
		}
	}
	return t.NewSpan(ctx, name)
}

func Logf(ctx context.Context, format string, arg ...interface{}) error {
	t := tracerFromContext(ctx)
	if t == nil {
		return nil
	}
	return t.Log(ctx, fmt.Sprintf(format, arg))
}

func Log(ctx context.Context, payload interface{}) error {
	t := tracerFromContext(ctx)
	if t == nil {
		return nil
	}
	return t.Log(ctx, payload)
}

func SetLabel(ctx context.Context, key string, value interface{}) {
	v := ctx.Value(labelsKey)
	var labels map[string]interface{}
	if v == nil {
		labels = make(map[string]interface{})
	} else {
		labels = v.(map[string]interface{})
	}
	labels[key] = value
}

func Finish(ctx context.Context) {
	t := tracerFromContext(ctx)
	if t == nil {
		return
	}
	v := ctx.Value(labelsKey)
	if v == nil {
		t.Finish(ctx, nil)
	} else {
		t.Finish(ctx, v.(map[string]interface{}))
	}
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
	traceKey  = contextKey("trace")
	labelsKey = contextKey("labels")
)
