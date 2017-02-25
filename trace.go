// Package trace defines common-use Dapper-style tracing APIs for the Go programming language.
//
// Package trace provides a backend-agnostic APIs and various tracing providers
// can be used with the package by importing various implementations of the Client interface.
package trace

import (
	"context"
	"runtime"
	"time"
)

// Client represents a client communicates with a tracing backend.
// Tracing backends are supposed to implement the interface in order to
// provide Go support.
//
// If you are not a backend provider, you will never have to interact with
// this interface directly.
type Client interface {
	// NewSpan creates a new span from the parent context's span.
	//
	// If parent context doesn't already have a span, it creates a top-level span.
	//
	// Span start time should be the given start.
	NewSpan(parent context.Context, name string) context.Context

	// Spans puts and existing span into a context derived from the given context.
	// Nil name and info is not allowed.
	Span(ctx context.Context, name string, info []byte) (context.Context, error)

	// Info returns the unique trace identifier assigned to the current context's trace tree.
	Info(ctx context.Context) []byte

	// Finish finishes the span in the context with the given labels.
	//
	// If causal is non-nil, span needs to be finalized with a causal relationship.
	// Nil labels should be accepted.
	//
	// Span should have given start and end time.
	Finish(ctx context.Context, causal []byte, labels map[string][]byte, start, end time.Time) error
}

// WithClient adds a Client into the current context later to be used to interact with
// the tracing backend.
//
// All trace package functions will do nothing if this function is not called with a non-nil trace client.
//
// TODO(jbd): Inject the client into the context or have a global registery?
func WithClient(ctx context.Context, c Client) context.Context {
	info := &traceInfo{
		client: c,
		labels: make(map[string][]byte),
	}
	return context.WithValue(ctx, traceInfoKey, info)
}

// Info returns the current context's span info. The format of info
// is specific to how tracing backend identifies a span in a trace tree.
//
// If context doesn't contain a span, it returns nil.
func Info(ctx context.Context) []byte {
	t := traceClientFromContext(ctx)
	if t == nil {
		return nil
	}
	return t.Info(ctx)
}

// FinishFunc finalizes the span from the current context.
// Each span context created by ChildSpan should be finished when their work is finished.
type FinishFunc func() error

// ChildSpan creates a new child span from the current context. Users are supposed to
// call Finish to finalize the span created by this function.
//
// If no name is given, caller function's name will be automatically.
//
// In a Dapper trace tree, the nodes are basic units of work represented as spans.
// If you need to represent any work indivually, you need to create a new span
// within the current context by calling this function.
// All the calls that is made by the returned span will be associated by the span created internally.
//
// If there is not trace client in the given context, ChildSpan does nothing.
func ChildSpan(ctx context.Context, name string) (context.Context, FinishFunc) {
	t := traceClientFromContext(ctx)
	if t == nil {
		return ctx, noop
	}
	if name == "" {
		// the name of the caller function
		pc, _, _, ok := runtime.Caller(1)
		if ok {
			fn := runtime.FuncForPC(pc)
			name = fn.Name()
		}
	}
	start := time.Now()
	newctx := t.NewSpan(ctx, name)
	finish := func() error {
		v := newctx.Value(traceInfoKey)
		if v == nil {
			return nil
		}
		return t.Finish(newctx, nil, v.(*traceInfo).labels, start, time.Now())
	}
	return newctx, finish
}

func CausalSpan(ctx context.Context, name string) (context.Context, FinishFunc) {
	panic("not implemented")
}

// Span puts a span identified by name and info in a context derived from the current
// context and returns it.
//
// It is useful for propagation where you retrieved information about a remote span
// and want to resurrect it locally to create child spans from it.
//
// Name shouldn't be empty, info shouldn't be nil.
func Span(ctx context.Context, name string, info []byte, start, end time.Time) (context.Context, error) {
	t := traceClientFromContext(ctx)
	if t == nil {
		return ctx, nil
	}
	return t.Span(ctx, name, info)
}

// StartEnd creates a span started at start and finished at end.
//
// Most users will use ChildSpan and CausalSpan. Only users who already add an trace
// entry for an already finished event should depend on this call.
func StartEnd(ctx context.Context, name string, start, end time.Time) (context.Context, FinishFunc) {
	t := traceClientFromContext(ctx)
	if t == nil {
		return ctx, nil
	}
	newctx := t.NewSpan(ctx, name)
	finish := func() error {
		v := newctx.Value(traceInfoKey)
		if v == nil {
			return nil
		}
		return t.Finish(newctx, nil, v.(*traceInfo).labels, start, end)
	}
	return newctx, finish
}

// SetLabel sets label identified with key on the current span.
//
// If context doesn't contain a trace client, SetLabel does nothing.
//
// TODO(jbd): Investigate the case for string keys and []byte values.
func SetLabel(ctx context.Context, key string, value []byte) {
	v := ctx.Value(traceInfoKey)
	if v == nil {
		return
	}
	info := v.(*traceInfo)
	info.labels[key] = value
}

func traceClientFromContext(ctx context.Context) Client {
	v := ctx.Value(traceInfoKey)
	if v == nil {
		return nil
	}
	return v.(*traceInfo).client
}

type contextKey string

var traceInfoKey = contextKey("trace")

type traceInfo struct {
	client Client
	labels map[string][]byte
}

func noop() error { return nil }
