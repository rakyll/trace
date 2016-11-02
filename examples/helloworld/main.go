package main

import (
	"context"
	"log"
	"runtime"
	"time"

	"github.com/rakyll/trace"
)

var t *trace.Client

func main() {
	ctx := context.Background()

	client, err := trace.NewClient(ctx, "jbd-gce")
	if err != nil {
		log.Fatal(err)
	}
	t = client
	f(ctx)
}

func f(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "")
	defer s.Finish()

	go a1(ctx)
	a2(ctx)
	a3(ctx)
}

func a1(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "")
	defer s.Finish()

	s.Logf("this is a format string, num goroutines: %v", runtime.NumGoroutine())
	time.Sleep(100 * time.Millisecond)
}

func a2(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "a2")
	defer s.Finish()

	time.Sleep(200 * time.Millisecond)
}

func a3(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "a3")
	defer s.Finish()

	time.Sleep(300 * time.Millisecond)
}
