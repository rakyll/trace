// Package trace defines common-use Dapper-style tracing APIs for the Go programming language.
//
// Package trace provides a backend-agnostic APIs and various tracing providers
// can be used with the package by importing various implementations of the Client interface.
package trace

import (
	"errors"
	"net/http"
	"time"
)

var client Client

// TODO(jbd): A big TODO, we probably don't want to set a global client.
func Configure(c Client) {
	client = c
}

// Span represents a work.
type Span struct {
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

// NewChild creates a child span from s with the given name.
// Created child span needs to be finished by calling the finishing function.
func (s *Span) NewChild(name string) (*Span, FinishFunc) {
	child := &Span{
		id:     client.NewSpan(s.id, nil),
		labels: make(map[string][]byte),
	}
	start := time.Now()
	fn := func() error {
		return client.Finish(child.id, name, child.labels, start, time.Now())
	}
	return child, fn
}

func (s *Span) ToHTTPReq(req *http.Request) (*http.Request, error) {
	hc, ok := client.(HTTPCarrier)
	if ok {
		return req, errors.New("not supported")
	}
	err := hc.ToHTTPReq(req, s.id)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Span) NewCausual(name string) (*Span, FinishFunc) {
	panic("not yet")
}

// RemoteSpan represents spans created by a remote server.
// A remote span is useful to create children and causual relationships
// to a span living remotely.
type RemoteSpan struct {
	id []byte
}

// ID returns the backend-specific global identifier of the span.
func (s *RemoteSpan) ID() []byte {
	return s.id
}

// NewChild creates a child span from s with the given name.
// Created child span needs to be finished by calling the finishing function.
func (s *RemoteSpan) NewChild(name string) (*Span, FinishFunc) {
	child := &Span{
		id:     client.NewSpan(s.id, nil),
		labels: make(map[string][]byte),
	}
	start := time.Now()
	fn := func() error {
		return client.Finish(child.id, name, child.labels, start, time.Now())
	}
	return child, fn
}

func (s *RemoteSpan) NewCausual(name string) (*Span, FinishFunc) {
	panic("not yet")
}

func FromHTTPReq(req *http.Request) (*RemoteSpan, error) {
	hc, ok := client.(HTTPCarrier)
	if ok {
		return nil, errors.New("not supported")
	}
	id, err := hc.FromHTTPReq(req)
	if err != nil {
		return nil, err
	}
	return &RemoteSpan{id: id}, nil
}

// HTTPCarrier represents a mechanism that can attach the tracing
// information into an HTTP request or extract it from one.
type HTTPCarrier interface {
	FromHTTPReq(req *http.Request) (id []byte, err error)
	ToHTTPReq(req *http.Request, id []byte) error
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

func NewSpan(name string) (*Span, FinishFunc) {
	span := &Span{
		id:     client.NewSpan(nil, nil),
		labels: make(map[string][]byte),
	}
	start := time.Now()
	fn := func() error {
		return client.Finish(span.id, name, span.labels, start, time.Now())
	}
	return span, fn
}

// FinishFunc finalizes the span from the current context.
// Each span context created by ChildSpan should be finished when their work is finished.
type FinishFunc func() error
