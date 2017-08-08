package traceutil

import (
	"context"
	"io"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"
)

// RoundTripResult is a set of timestamps set at various stages of an outgoing
// HTTP request intended to be passed to a FinishFunc passed in to RoundTrip.
type RoundTripResult struct {
	Started              time.Time
	WroteHeaders         time.Time
	WroteRequest         time.Time
	GotFirstResponseByte time.Time
	ReadHeaders          time.Time
	ReadResponse         time.Time

	BytesWritten int // TODO: implemenent
	BytesRead    int

	// The request traced. It is not safe to modify nor read from its body
	// or headers.
	Request *http.Request

	// The response traced. It is not safe to modify nor read from its body
	// of headers.  It is nil if no response was recieved.
	Response *http.Response

	Err error
}

func propagateHTTP(req *http.Request) {
	// TODO: implement
}

type traceRoundTripper struct {
	inner    http.RoundTripper
	finished FinishFunc
}

// RoundTrip returns a http.RoundTripper that wraps inner for purposes of
// recording timestamps to a RoundTripResult which is passed to finished on
// completion or error of the returned http.RoundTripper.
func RoundTrip(inner http.RoundTripper, finished FinishFunc) http.RoundTripper {
	return traceRoundTripper{inner: inner, finished: finished}
}

type FinishFunc func(result *RoundTripResult)

// FinishBasic records a single span for the request representing the from
// first request byte written to last response byte read.
func FinishBasic() FinishFunc {
	panic("TODO")
}

// FinishDetail records sevaral detailed spans for the request representing the
// from first request byte written to last response byte read.
//
// The spans are:
//
//    |==RoundTrip=======================================================|
//    |==WroteHeader==                                                   |
//    |               ==WroteBody==                                      |
//    |                                  ==ReadFirstByte==               |
//    |                                                   ==ReadHeaders==|
//    |==================================================================|
//
func FinishDetail() FinishFunc {
	panic("TODO")
}

func (r traceRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var (
		trace   = newClientTrace()
		ctx     = httptrace.WithClientTrace(req.Context(), trace.client)
		timeNow = timeNowFromContext(ctx)
	)
	req = req.WithContext(ctx)

	// TODO(bmizerany): check if already configured for propagation as to
	// not override anything custom? We could do this with a key in the
	// req.Context.
	propagateHTTP(req)

	trace.Request = req
	trace.Started = timeNow()

	res, err := r.inner.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	trace.Response = res
	trace.ReadHeaders = timeNow()

	finish := func(err error) {
		if trace.Err != nil {
			trace.Err = err
		}
		trace.ReadResponse = timeNow()
		if r.finished != nil {
			// TODO: doc you don't own this after done
			r.finished(trace.RoundTripResult)
		}
		clientTracePool.Put(trace)
	}

	res.Body = traceBody{
		body:   res.Body,
		result: trace.RoundTripResult,
		finish: finish,
	}
	return res, nil
}

type traceBody struct {
	body   io.ReadCloser
	result *RoundTripResult
	finish func(error)
}

func (t traceBody) Read(b []byte) (int, error) {
	n, err := t.body.Read(b)
	t.result.BytesRead += n
	if err != nil {
		if err == io.EOF {
			err = nil
		}
		t.finish(err)
	}
	return n, err
}

func (t traceBody) Close() error {
	err := t.body.Close()
	t.finish(err)
	return err
}

type clientTrace struct {
	*RoundTripResult
	client *httptrace.ClientTrace
}

var clientTracePool sync.Pool

func newClientTrace() *clientTrace {
	trace, ok := clientTracePool.Get().(*clientTrace)
	if ok {
		*trace.RoundTripResult = RoundTripResult{}
		return trace
	}

	trace = &clientTrace{
		RoundTripResult: new(RoundTripResult),
	}

	trace.client = &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			trace.GotFirstResponseByte = time.Now()
		},
		WroteHeaders: func() {
			trace.WroteHeaders = time.Now()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			trace.WroteRequest = time.Now()

			// TODO: do something with info.Err?
			//
			// NOTE: BE CAREFUL - it can race with
			// onErrBody since we may start reading the
			// response before all of the request is
			// written.
		},
	}
	return trace
}

type timeNowKey struct{}

func timeNowFromContext(ctx context.Context) func() time.Time {
	v := ctx.Value(timeNowKey{})
	if v != nil {
		return v.(func() time.Time)
	}
	return time.Now
}
