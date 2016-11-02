package main

import (
	"context"
	"log"
	"time"

	"github.com/rakyll/trace"
)

var t *trace.Client

func main() {
	ctx := context.Background()

	client, err := trace.NewClient(ctx, "bamboo-shift-504")
	if err != nil {
		log.Fatal(err)
	}
	t = client
	f(ctx)
}

func f(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "")
	defer s.Finish()

	a1(ctx)
	a2(ctx)
	a3(ctx)
}

func a1(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "")
	defer s.Finish()

	s.Logf("something bad is happening at %v", time.Now())
	time.Sleep(1 * time.Second)
}

func a2(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "a2")
	defer s.Finish()

	time.Sleep(2 * time.Second)
}

func a3(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "a3")
	defer s.Finish()

	time.Sleep(3 * time.Second)
}
