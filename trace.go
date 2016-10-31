package trace

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	muParents sync.Mutex
	parents   = make([]*Span, 0)
)

func Log() {
	muParents.Lock()
	defer muParents.Unlock()

	logSpans(parents, 0)
}

func logSpans(spans []*Span, indent int) {
	for _, p := range spans {
		fmt.Printf("%v%v\n", strings.Repeat("-", 2*indent), p.name)
		logSpans(p.spans, indent+1)
	}
}

type Span struct {
	name string

	mu    sync.Mutex
	spans []*Span
	begin time.Time
	end   time.Time
}

func NewSpan(ctx context.Context, name string) (context.Context, *Span) {
	parent := ContextSpan(ctx)
	s := &Span{
		name:  name,
		spans: make([]*Span, 0),
		begin: time.Now(),
	}
	c := context.WithValue(ctx, ctxKey, s)
	if parent != nil {
		parent.mu.Lock()
		parent.spans = append(parent.spans, s)
		parent.mu.Unlock()
	} else {
		muParents.Lock()
		parents = append(parents, s)
		muParents.Unlock()
	}
	return c, s
}

func ContextSpan(ctx context.Context) *Span {
	v := ctx.Value(ctxKey)
	if v == nil {
		return nil
	}
	return ctx.Value(ctxKey).(*Span)
}

func (s *Span) Finish() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.end = time.Now()
}

var ctxKey = struct{}{}
