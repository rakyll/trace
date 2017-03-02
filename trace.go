// Package trace defines common-use Dapper-style tracing APIs for the Go programming language.
//
// Package trace provides a backend-agnostic APIs and various tracing providers
// can be used with the package by importing various implementations of the Client interface.
package trace

import (
	"net/http"
	"time"
)

// Span represents a work.
type Span struct {
	client Client

	id     []byte
	labels map[string][]byte
}

// ID returns the backend-specific global identifier of the span.
func (s *Span) ID() []byte {
	return s.id
}

// SetLabel allows you to set a label on the current span. Labels are
// arbitary information you want to collect in the lifetime of a span.
func (s *Span) SetLabel(key string, val []byte) {
	s.labels[key] = val
}

// Child creates a child span from s with the given name.
// Created child span needs to be finished by calling the finishing function.
func (s *Span) Child(name string) (*Span, FinishFunc) {
	child := &Span{
		client: s.client,
		id:     s.client.NewSpan(s.id, nil),
		labels: make(map[string][]byte),
	}
	start := time.Now()
	fn := func() error {
		return s.client.Finish(child.id, name, child.labels, start, time.Now())
	}
	return child, fn
}

func (s *Span) Causual(name string) (*Span, FinishFunc) {
	panic("not yet")
}

// RemoteSpan represents spans created by a remote server.
// A remote span is useful to create children and causual relationships
// to a span living remotely.
type RemoteSpan struct {
	client Client
	id     []byte
}

// ID returns the backend-specific global identifier of the span.
func (s *RemoteSpan) ID() []byte {
	return s.id
}

// Child creates a child span from s with the given name.
// Created child span needs to be finished by calling the finishing function.
func (s *RemoteSpan) Child(name string) (*Span, FinishFunc) {
	child := &Span{
		client: s.client,
		id:     s.client.NewSpan(s.id, nil),
		labels: make(map[string][]byte),
	}
	start := time.Now()
	fn := func() error {
		return s.client.Finish(child.id, name, child.labels, start, time.Now())
	}
	return child, fn
}

func (s *RemoteSpan) Causual(name string) (*Span, FinishFunc) {
	panic("not yet")
}

// HTTPCarrier represents a mechanism that can attach the tracing
// information into an HTTP request or extract it from one.
type HTTPCarrier interface {
	FromHTTP(req *http.Request) (id []byte, err error)
	ToHTTP(req *http.Request, id []byte) error
}

// Client represents a client communicates with a tracing backend.
// Tracing backends are supposed to implement the interface in order to
// provide Go support.
//
//
// A Client is an HTTPCarrier if it can propagate the tracing
// information via an HTTP request.
//
// If you are not a tracing provider, you will never have to interact with
// this interface directly.
type Client interface {
	// NewSpan creates a new span from the parent context's span.
	//
	// If parent context doesn't already have a span, it creates a top-level span.
	//
	// Span start time should be the given start.
	NewSpan(parent []byte, causal []byte) (id []byte)

	// Finish finishes the span in the context with the given labels.
	//
	// If causal is non-nil, span needs to be finalized with a causal relationship.
	// Nil labels should be accepted.
	//
	// Span should have given start and end time.
	Finish(id []byte, name string, labels map[string][]byte, start, end time.Time) error
}

// FinishFunc finalizes the span from the current context.
// Each span context created by ChildSpan should be finished when their work is finished.
type FinishFunc func() error

// NewTrace is the entry point
func NewTrace(c Client) *Span {
	return &Span{
		client: c,
		id:     nil, // root-span
	}
}
