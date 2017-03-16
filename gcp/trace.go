package gcp

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
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

type span struct {
	id   spanID
	name string

	labels map[string][]byte
	start  time.Time
	end    time.Time
}

func finish(c *client, projID string, spans []*span) error {
	// group by trace ID.
	var traces []*api.Trace
	for _, s := range spans {
		t := constructTrace(projID, s)
		traces = append(traces, t)
	}
	return c.upload(traces)
}

func constructTrace(projID string, span *span) *api.Trace {
	s := &api.TraceSpan{
		Name:         span.name,
		SpanId:       span.id.ID,
		ParentSpanId: span.id.ParentID,
		StartTime:    span.start.In(time.UTC).Format(time.RFC3339Nano),
		EndTime:      span.end.In(time.UTC).Format(time.RFC3339Nano),
		// TODO(jbd): add labels
	}
	return &api.Trace{
		ProjectId: projID,
		TraceId:   span.id.TraceID,
		Spans:     []*api.TraceSpan{s},
	}
}

type spanID struct {
	TraceID  string
	ParentID uint64
	ID       uint64
}
