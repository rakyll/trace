// Package trace defines common-use Dapper-style tracing APIs for the Go programming language.
package trace

import (
	"context"
	"sync"
	"time"
)

var (
	notifierMu sync.RWMutex
	notifier   func(s *Span)
)

type Span struct {
	TraceID  []byte
	ParentID []byte
	ID       []byte

	Name      string
	StartTime time.Time
	EndTime   time.Time

	labelsMu sync.Mutex
	labels   map[string]string
}

func NewSpan(name string) *Span {
	notifierMu.RLock()
	defer notifierMu.RUnlock()
	if notifier == nil {
		return nil
	}

	// TODO(jbd): Return nil if not sampled.
	return &Span{
		TraceID:   nil, //TODO(jbd): Generate.
		ID:        nil, //TODO(jbd): Generate.
		Name:      name,
		StartTime: time.Now(),
	}
}

func (s *Span) End() {
	if s == nil {
		return
	}
	s.EndTime = time.Now()
	notifierMu.RLock()
	defer notifierMu.RUnlock()

	if notifier != nil {
		notifier(s)
	}
}

func (s *Span) NewChild(name string) *Span {
	if s == nil {
		return nil
	}
	return &Span{
		TraceID:   s.TraceID,
		ParentID:  s.ID,
		ID:        nil, //TODO(jbd): Generate.
		Name:      name,
		StartTime: time.Now(),
	}
}

func (s *Span) SetLabels(args ...string) {
	if len(args)%2 != 0 {
		panic("even number of arguments required")
	}
	s.labelsMu.Lock()
	defer s.labelsMu.Unlock()

	for i := 0; i < len(args); i += 2 {
		s.labels[args[i]] = s.labels[args[i+1]]
	}
}

func (s *Span) ForLabels(f func(key, value string)) {
	for k, v := range s.labels {
		f(k, v)
	}
}

func NewContext(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, ctxKey, span)
}

func FromContext(ctx context.Context) *Span {
	return ctx.Value(ctxKey).(*Span)
}

func Start(fn func(s *Span)) {
	notifierMu.Lock()
	defer notifierMu.Unlock()

	notifier = fn
}

func Stop() {
	notifierMu.Lock()
	defer notifierMu.Unlock()

	notifier = nil
}

type spanContextKey struct{}

var ctxKey = spanContextKey{}
