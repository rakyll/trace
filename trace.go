// Package trace defines common-use Dapper-style tracing APIs for the Go programming language.
//
// Package trace provides a backend-agnostic APIs and various tracing providers
// can be used with the package by importing various implementations of the Client interface.
package trace

import (
	"context"
	"fmt"
	"runtime"
)

// Client represents a client communicates with a tracing backend.
// Tracing backends are supposed to implement the interface in order to
// provide Go support.
//
// If you are not a backend provider, you will never have to interact with
// this interface directly.
type Client interface {
	// NewSpan creates a new child span from the current span in the current context.
	// If there are no current spans in the current span, a top-level span is created.
	NewSpan(ctx context.Context, name string) context.Context

	// TraceID returns the unique trace identifier assigned to the current context's trace tree.
	TraceID(ctx context.Context) []byte

	// Finish finishes the span in the context with the given labels. Nil labels
	// should be accepted.
	Finish(ctx context.Context, labels map[string]interface{}) error
}

// TODO(jbd): Add logger to Client?

// WithClient adds a Client into the current context later to be used to interact with
// the tracing backend.
//
// All trace package functions will act as no-ops if this function is not called with a non-nil trace client.
func WithClient(ctx context.Context, c Client) context.Context {
	return context.WithValue(ctx, traceKey, c)
}

// TraceID returns the current context's unique trace tree ID.
//
// If context doesn't contain a tracer, it returns empty string.
func TraceID(ctx context.Context) []byte {
	t := tracerFromContext(ctx)
	if t == nil {
		return nil
	}
	return t.TraceID(ctx)
}

// FinishFunc finalizes the span from the current context.
// Each span context created by WithSpan should be finished when their work is finished.
type FinishFunc func() error

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
func WithSpan(ctx context.Context, name string) (context.Context, FinishFunc) {
	t := tracerFromContext(ctx)
	if t == nil {
		noop := func() error { return nil }
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
	newctx := t.NewSpan(ctx, name)
	finish := func() error {
		// TODO(jbd): pass the labels.
		v := newctx.Value(labelsKey)
		if v == nil {
			return t.Finish(newctx, nil)
		}
		return t.Finish(newctx, v.(map[string]interface{}))
	}
	return newctx, finish
}

type stringer struct {
	format string
	args   []interface{}
}

func (s *stringer) String() string {
	return fmt.Sprintf(s.format, s.args...)
}

type Logger interface {
	Log(ctx context.Context, arg ...fmt.Stringer) error
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

func tracerFromContext(ctx context.Context) Client {
	v := ctx.Value(traceKey)
	if v == nil {
		return nil
	}
	return v.(Client)
}

type contextKey string

var (
	traceKey  = contextKey("trace")
	labelsKey = contextKey("labels")
)
