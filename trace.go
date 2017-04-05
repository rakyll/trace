// Package trace defines common-use Dapper-style tracing APIs for the Go programming language.
//
// Package trace provides a backend-agnostic APIs and various tracing providers
// can be used with the package by importing various implementations of the Client interface.
package trace

import (
	"context"
	"net/http"
	"time"

	"github.com/rakyll/trace/minitrace"
)

var client Client

// TODO(jbd): how to propagate a subset of labels?

// TODO(jbd): should we support a span to have multiple parents?

// TODO(jbd): set error/state on finish.

// TODO(jbd): Avoid things if client = nil.

// TODO(jbd): A big TODO, we probably don't want to set a global client.
func Configure(c Client) {
	client = c
}

// Span represents a work.
type Span struct {
	ID          []byte            // represents the global identifier of the span.
	Annotations map[string][]byte // annotations set on this span.
}

// Annotate allows you to attach data to a span. Key-value pairs are
// arbitary information you want to collect in the lifetime of a span.
func (s *Span) Annotate(key string, val []byte) {
	if s == nil {
		return
	}
	s.Annotations[key] = val
}

// Child creates a child span from s with the given name.
// Created child span needs to be finished by calling
// the finishing function.
func (s *Span) Child(name string, linked ...*Span) (*Span, FinishFunc) {
	if s == nil {
		return nil, noop
	}
	child := &Span{
		ID:          client.NewSpan(s.ID),
		Annotations: make(map[string][]byte),
	}
	start := time.Now()
	fn := func() error {
		return client.Finish(child.ID, name, spanIDs(linked), child.Annotations, start, time.Now())
	}
	return child, fn
}

// ToHTTPReq injects the span information in the given request
// and returns the modified request.
//
// If the current client is not supporting HTTP propagation,
// an error is returned.
func (s *Span) ToHTTPReq(req *http.Request) (*http.Request, error) {
	if s == nil {
		return req, nil
	}
	hc, ok := client.(minitrace.HTTPInjector)
	if !ok {
		return req, nil
	}
	return hc.SpanToReq(req, &sspan{id: s.ID})
}

type sspan struct {
	id []byte
}

func (s *sspan) ID() []byte {
	return s.id
}

func (s *sspan) Annotations() []byte {
	return nil
}

// FromHTTPReq creates a *Span from an incoming request.
//
// An error will be returned if the current tracing client is
// not supporting propagation via HTTP.
func FromHTTPReq(req *http.Request) (*Span, error) {
	hc, ok := client.(minitrace.HTTPExtractor)
	if !ok {
		return nil, nil
	}
	span, err := hc.SpanFromReq(req)
	if err != nil {
		return nil, err
	}
	return &Span{ID: span.ID()}, nil
}

// Client represents a client communicates with a tracing backend.
// Tracing backends are supposed to implement the interface in order to
// provide Go support.
//
// A Client is an HTTPExtractor if it can extract spans from an incoming HTTP request.
// A Client in an HTTPInjector if it can inject span info into an outgoing HTTP request.
//
// If you are not a tracing provider, you will never have to interact with
// this interface directly.
type Client interface {
	NewSpan(parent []byte) (id []byte)
	Finish(id []byte, name string, linked [][]byte, annotations map[string][]byte, start, end time.Time) error
}

// NewSpan creates a new root-level span.
//
// The span must be finished when the job it represents it is finished.
func NewSpan(name string, linked ...*Span) (*Span, FinishFunc) {
	span := &Span{
		ID:          client.NewSpan(nil),
		Annotations: make(map[string][]byte),
	}
	start := time.Now()
	fn := func() error {
		return client.Finish(span.ID, name, spanIDs(linked), span.Annotations, start, time.Now())
	}
	return span, fn
}

// NewContext returns a context with the span in.
func NewContext(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanKey, span)
}

// FromContext returns a span from the given context.
func FromContext(ctx context.Context) *Span {
	return ctx.Value(spanKey).(*Span)
}

func spanIDs(spans []*Span) [][]byte {
	var ids [][]byte
	for _, s := range spans {
		ids = append(ids, s.ID)
	}
	return ids
}

// FinishFunc finalizes its span.
type FinishFunc func() error

type contextKey struct{}

var spanKey = contextKey{}

func noop() error { return nil }
