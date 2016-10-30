package trace

import (
	"log"
	"time"
)

var Logger log.Logger

type Span struct {
	name  string
	spans []*Span

	begin time.Time
	end   time.Time
}

func NewSpan(name string) *Span {
	return &Span{
		name:  name,
		spans: make([]*Span, 0),
		begin: time.Now(),
	}
}

func (s *Span) NewSpan(name string) *Span {
	child := NewSpan(name)
	s.spans = append(s.spans, child)
	return child
}

func (s *Span) Finish() {
	s.end = time.Now()
}
