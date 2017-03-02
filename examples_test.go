package trace_test

import (
	"log"

	"github.com/rakyll/trace"
	"github.com/rakyll/trace/gcp"
)

func ExampleNewTrace() {
	c, err := gcp.NewClient("project-id")
	if err != nil {
		log.Fatal(err)
	}

	root := trace.NewTrace(c)
	sp, finish := root.Child("/whatever")
	defer finish()

	sp.SetLabel("hello", []byte("error happened"))

	trace.ToHTTP(c, req, sp)

	carrier, ok := c.(HTTPCarrier)
	if ok {
		carrier.ToHTTP(req, sp.ID())
	}
}
