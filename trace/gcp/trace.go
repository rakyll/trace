package gcp

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	api "google.golang.org/api/cloudtrace/v1"
)

func init() {
	// Set spanIDCounter and spanIDIncrement to random values.  nextSpanID will
	// return an arithmetic progression using these values, skipping zero.  We set
	// the LSB of spanIDIncrement to 1, so that the cycle length is 2^64.
	binary.Read(rand.Reader, binary.LittleEndian, &spanIDCounter)
	binary.Read(rand.Reader, binary.LittleEndian, &spanIDIncrement)
	spanIDIncrement |= 1
}

var (
	spanIDCounter   uint64
	spanIDIncrement uint64
)

// nextSpanID returns a new span ID.  It will never return zero.
func nextSpanID() uint64 {
	var id uint64
	for id == 0 {
		id = atomic.AddUint64(&spanIDCounter, spanIDIncrement)
	}
	return id
}

// nextTraceID returns a new trace ID.
func nextTraceID() string {
	id1 := nextSpanID()
	id2 := nextSpanID()
	return fmt.Sprintf("%016x%016x", id1, id2)
}

type trace struct {
	id string

	sync.Mutex
	spans []*span
}

func (t *trace) finish(ctx context.Context, s *span) error {
	s.end = time.Now()
	c := contextClient(ctx)
	return c.upload([]*api.Trace{t.constructTrace(c.proj, t.spans)})
}

func (t *trace) constructTrace(projID string, spans []*span) *api.Trace {
	apiSpans := make([]*api.TraceSpan, len(spans))
	for i, sp := range spans {
		s := &api.TraceSpan{
			Name:         sp.name,
			SpanId:       sp.id,
			ParentSpanId: sp.parentID,
			StartTime:    sp.start.In(time.UTC).Format(time.RFC3339Nano),
			EndTime:      sp.end.In(time.UTC).Format(time.RFC3339Nano),
			// TODO(jbd): add labels
		}
		apiSpans[i] = s
	}
	return &api.Trace{
		ProjectId: projID,
		TraceId:   t.id,
		Spans:     apiSpans,
	}
}

type span struct {
	client *Client
	trace  *trace

	parentID uint64
	id       uint64
	name     string

	start time.Time
	end   time.Time
}

func contextClient(ctx context.Context) *Client {
	v := ctx.Value(clientKey)
	if v == nil {
		return nil
	}
	return v.(*Client)
}

// contextSpan returns a span from the current context or nil
// if no contexts exists in the current context.
func contextSpan(ctx context.Context) *span {
	v := ctx.Value(spanKey)
	if v == nil {
		return nil
	}
	return v.(*span)
}

type key string

var spanKey = key("span")
var clientKey = key("client")
