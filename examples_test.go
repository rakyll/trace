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
		ctx = trace.WithSpan(ctx, "")
		defer trace.Finish(ctx)

		time.Sleep(time.Minute)
		trace.Logf(ctx, "it took too long...")
	}

	ctx = trace.WithClient(context.Background(), tc)
	call(ctx)
}

func ExampleFinish() {
	ctx = trace.WithSpan(ctx, "") // Creates new span for the context.
	defer trace.Finish(ctx)
}

func ExampleWithSpan() {
	// Create a span that will track the function that
	// reads the users from the users service.
	ctx = trace.WithSpan(ctx, "/api.ReadUsers")
	defer trace.Finish(ctx)
}
