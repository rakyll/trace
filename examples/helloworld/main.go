package main

import (
	"context"
	"log"

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
	ctx, s := t.NewSpan(ctx, "f")
	defer s.Finish()

	a1(ctx)
	a2(ctx)
	a3(ctx)
}

func a1(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "a")
	defer s.Finish()
}

func a2(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "a")
	defer s.Finish()
}

func a3(ctx context.Context) {
	ctx, s := t.NewSpan(ctx, "a")
	defer s.Finish()
}
