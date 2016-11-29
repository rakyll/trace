// Package trace defines common-use Dapper-style tracing APIs for the Go programming language.
package trace

import (
	"context"
	"fmt"
	"runtime"
)

// TraceClient represents a client communicates with a tracing backend.
// Tracing backends are supposed to implement the interface in order to
// provide Go support.
//
// If you are not a backend provider, you will never have to interact with
// this interface directly.
type TraceClient interface {
	// NewSpan creates a new child span from the current span in the current context.
	// If there are no current spans in the current span, a top-level span is created.
	NewSpan(ctx context.Context, name string) context.Context

	// TraceID returns the unique trace ID assigned to the current context's trace tree.
	TraceID(ctx context.Context) string

	// Finish finishes the span in the context with the given labels. Nil labels
	// should be accepted.
	Finish(ctx context.Context, labels map[string]interface{})

	// Log associates the payload with the span in the context and logs it.
	Log(ctx context.Context, payload interface{}) error
}

// WithTrace adds a TraceClient into the current context later to be used to interact with
// the tracing backend.
//
// All trace package functions will act as no-ops if this function is not called with a non-nil trace client.
func WithTrace(ctx context.Context, t TraceClient) context.Context {
	return context.WithValue(ctx, traceKey, t)
}

// TraceID returns the current context's unique trace tree ID.
//
// If context doesn't contain a tracer, it returns empty string.
func TraceID(ctx context.Context) string {
	t := tracerFromContext(ctx)
	if t == nil {
		return ""
	}
	return t.TraceID(ctx)
}

// WithSpan creates a new child span from the current context. Users are supposed to
// call Finish to finalize the span created by this function.
//
// If no name is given, caller function's name will be automatically.
//
// In a Dapper trace tree, the nodes are basic units of work represented as spans.
// If you need to represent any work indivually, you need to create a new span
// within the current context by calling this function.
// All the calls that is made by the returned span will be associated by the span created internally.
//
// If there is not trace client in the given context, WithSpan does nothing.
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

// Logf is like log.Printf.
// It formats the given string, associates it with the context span and logs it at the backend.
//
// If context doesn't contain a trace client, Logf acts like as a no-op function.
func Logf(ctx context.Context, format string, arg ...interface{}) error {
	t := tracerFromContext(ctx)
	if t == nil {
		return nil
	}
	return t.Log(ctx, fmt.Sprintf(format, arg))
}

// Log associates the given payload with the span in the current context and logs it at the backend.
//
// If context doesn't contain a trace client, Log acts like as a no-op function.
func Log(ctx context.Context, payload interface{}) error {
	t := tracerFromContext(ctx)
	if t == nil {
		return nil
	}
	return t.Log(ctx, payload)
}

// SetLabel sets label identified with key on the current span.
//
// If context doesn't contain a trace client, SetLabel acts like as a no-op function.
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

// Finish finalizes the span from the current context.
// Each span context created by WithSpan should be finished when their work is finished.
//
// If context doesn't contain a trace client, Finish acts like as a no-op function.
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

func tracerFromContext(ctx context.Context) TraceClient {
	v := ctx.Value(traceKey)
	if v == nil {
		return nil
	}
	return v.(TraceClient)
}

type contextKey string

var (
	traceKey  = contextKey("trace")
	labelsKey = contextKey("labels")
)
