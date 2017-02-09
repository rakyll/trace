package main

import (
	"context"
	"log"
	"time"

	"github.com/rakyll/trace"
	"github.com/rakyll/trace/gcp"
)

func main() {
	ctx := context.Background()
	c, err := gcp.NewClient(ctx, "jbd-gce")
	if err != nil {
		log.Fatal(err)
	}

	ctx = trace.WithClient(ctx, c)
	f(ctx)
}

func f(ctx context.Context) {
	ctx, finish := trace.ChildSpan(ctx, "")
	defer finish()

	go a1(ctx)
	a2(ctx)
	a3(ctx)
}

func a1(ctx context.Context) {
	ctx, finish := trace.ChildSpan(ctx, "")
	defer finish()

	time.Sleep(100 * time.Millisecond)
}

func a2(ctx context.Context) {
	ctx, finish := trace.ChildSpan(ctx, "a2")
	defer finish()

	time.Sleep(200 * time.Millisecond)
}

func a3(ctx context.Context) {
	ctx, finish := trace.ChildSpan(ctx, "a3")
	defer finish()

	time.Sleep(300 * time.Millisecond)
}
