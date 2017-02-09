package trace_test

import (
	"context"
	"log"
	"net/http"

	"github.com/rakyll/trace"
)

var ctx = context.Background()
var tc = trace.Client(nil)

func Example() {
	call := func(ctx context.Context) {
		ctx, finish := trace.ChildSpan(ctx, "/getUsers")
		defer finish()

		// Do the actual call to get users.
		_ = ctx
	}

	ctx = trace.WithClient(context.Background(), tc)
	call(ctx)
}

func ExampleSpan() {
	ctx := context.Background()

	http.HandleFunc("/getUserStatuses", func(w http.ResponseWriter, req *http.Request) {
		info := req.Header.Get("X-Trace")

		// Resurrect an already existing span created by a remote server
		// to create a child span from it.
		ctx, err := trace.Span(ctx, "/getUsers", []byte(info))
		if err != nil {
			log.Fatalf("Cannot resurrect a remote span: %v", err)
		}

		ctx, finish := trace.ChildSpan(ctx, "/getUserStatuses")
		defer finish()

		// Do the actual call to get user statuses.
		_ = ctx
	})
}

func ExampleFinishFunc() {
	ctx, finish := trace.ChildSpan(ctx, "/getUsers") // Creates new span for the context.
	defer finish()

	_ = ctx // keep using ctx
}

func ExampleChildSpan() {
	// Create a span that will track the function that
	// reads the users from the users service.
	ctx, finish := trace.ChildSpan(ctx, "/getUsers")
	defer finish()

	_ = ctx // keep using ctx
}
