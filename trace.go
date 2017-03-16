// Package trace defines common-use Dapper-style tracing APIs for the Go programming language.
//
// Package trace provides a backend-agnostic APIs and various tracing providers
// can be used with the package by importing various implementations of the Client interface.
package trace

import (
	"context"
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

// Child creates a child span from s with the given name.
// Created child span needs to be finished by calling
// the finishing function.
func (s *Span) Child(name string) (*Span, FinishFunc) {
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

// ToHTTPReq injects the span information in the given request
// and returns the modified request.
//
// If the current client is not supporting HTTP propagation,
// an error is returned.
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

func (s *Span) Causal(name string) (*Span, FinishFunc) {
	panic("not yet")
}

// FromHTTPReq creates a *Span from an incoming request.
//
// An error will be returned if the current tracing client is
// not supporting propagation via HTTP.
func FromHTTPReq(req *http.Request) (*Span, error) {
	hc, ok := client.(HTTPCarrier)
	if ok {
		return nil, errors.New("not supported")
	}
	id, err := hc.FromHTTPReq(req)
	if err != nil {
		return nil, err
	}
	return &Span{id: id}, nil
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
// A Client is an HTTPCarrier if it can propagate the tracing
// information via an HTTP request.
//
// If you are not a tracing provider, you will never have to interact with
// this interface directly.
type Client interface {
	NewSpan(parent []byte, causal []byte) (id []byte)
	Finish(id []byte, name string, labels map[string][]byte, start, end time.Time) error
}

// NewSpan creates a new root-level span.
//
// The span must be finished when the job it represents it is finished.
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

func NewContext(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanKey, span)
}

func FromContext(ctx context.Context) *Span {
	return ctx.Value(spanKey).(*Span)
}

// FinishFunc finalizes the span from the current context.
// Each span context created by ChildSpan should be finished when their work is finished.
type FinishFunc func() error

type contextKey struct{}

var spanKey = contextKey{}
