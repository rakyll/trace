package trace

import "time"

type Span struct {
	name string

	begin time.Time
	end   time.Time
}

func (s *Span) Finish() {
	panic("not implemented")
}

type Trace struct {
	name  string
	spans []Span
}
