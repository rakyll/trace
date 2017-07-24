package trace_test

import (
	"fmt"
	"time"

	"github.com/rakyll/trace"
)

func Example() {
	trace.Start(func(s *trace.Span) {
		fmt.Printf("recorded span: %v\n", s)
	})

	s := trace.NewSpan("/foo")
	time.Sleep(time.Second)
	s.End()

	time.Sleep(time.Second)
}
