package minitrace

import (
	"context"
	"net/http"
)

type Span []byte

func NewContext(ctx context.Context, s Span) context.Context {
	return context.WithValue(ctx, spanKey, s)
}

func FromContext(ctx context.Context) Span {
	return ctx.Value(spanKey).(Span)
}

type HTTPCarrier interface {
	SpanFromReq(req *http.Request) (Span, error)
	SpanToReq(req *http.Request, s Span) error
}

type contextKey struct{}

var spanKey = contextKey{}
