package trace_test

import (
	"context"
	"time"

	"github.com/rakyll/trace"
)

var ctx = context.Background()
var tc = trace.Client(nil)

func Example() {
	call := func(ctx context.Context) {
		ctx, finish := trace.WithSpan(ctx, "")
		defer finish()

		time.Sleep(time.Minute)
	}

	ctx = trace.WithClient(context.Background(), tc)
	call(ctx)
}

func ExampleFinishFunc() {
	ctx, finish := trace.WithSpan(ctx, "") // Creates new span for the context.
	defer finish()

	_ = ctx // keep using ctx
}

func ExampleWithSpan() {
	// Create a span that will track the function that
	// reads the users from the users service.
	ctx, finish := trace.WithSpan(ctx, "/api.ReadUsers")
	defer finish()

	_ = ctx // keep using ctx
}
