package trace_test

import (
	"context"
	"log"

	"github.com/rakyll/gcptrace/trace"
	"github.com/rakyll/gcptrace/trace/gcp"
)

func Example() {
	call := func(ctx context.Context) {
		ctx = trace.WithSpan(ctx, "")
		defer trace.Finish(ctx)

		trace.Logf(ctx, "it took too long...")
	}

	ctx := context.Background()
	c, err := gcp.NewClient(ctx, "jbd-gce")
	if err != nil {
		log.Fatal(err)
	}

	ctx = trace.WithTrace(ctx, c)
	call(ctx)
}
