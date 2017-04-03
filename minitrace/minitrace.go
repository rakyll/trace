// Package minitrace contains primitives to support
// propagation of tracing information.
package minitrace

// TODO(jbd): Rename this package to trace and propose it to the standard library.

// TODO(jbd): Rename Span to SpanID? Then, does it represent propagated labels?

import (
	"context"
	"net/http"
)

// Span represents a unit of work.
//
// ID identifies a span globally in a tracing system.
//
// Annotations return the labels propagated with the span.
// Encoding method might be tracing-backend specific.
type Span interface {
	ID() []byte
	Annotations() []byte
}

// NewContext returns a context with the given span.
func NewContext(ctx context.Context, s Span) context.Context {
	return context.WithValue(ctx, spanKey, s)
}

// FromContext returns the span from the context.
func FromContext(ctx context.Context) Span {
	return ctx.Value(spanKey).(Span)
}

// HTTPCarrier allows spans to be propagated via HTTP requests.
//
// SpanFromReq returns a span from the incoming HTTP request.
// If the incoming request doesn't contain trace information,
// a nil span is returned with no errors.
//
// SpanToReq mutates the outgoing request with the span, and
// returns a shallow copy of the request. If span is nil,
// req is not mutated.
type HTTPCarrier interface {
	SpanFromReq(req *http.Request) (Span, error)
	SpanToReq(req *http.Request, s Span) (*http.Request, error)
}

type contextKey struct{}

var spanKey = contextKey{}
