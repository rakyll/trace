package main

import (
	"context"

	"github.com/rakyll/trace"
)

func main() {
	f(context.TODO())
}

func f(ctx context.Context) {
	ctx, s := trace.NewSpan(ctx, "f")
	defer s.Finish()

	a1(ctx)
	a2(ctx)
	a3(ctx)

	trace.Log()
}

func a1(ctx context.Context) {
	ctx, s := trace.NewSpan(ctx, "a")
	defer s.Finish()
}

func a2(ctx context.Context) {
	ctx, s := trace.NewSpan(ctx, "a")
	defer s.Finish()
}

func a3(ctx context.Context) {
	ctx, s := trace.NewSpan(ctx, "a")
	defer s.Finish()
}
