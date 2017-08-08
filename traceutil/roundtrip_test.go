package traceutil

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func BenchmarkRoundTrip(b *testing.B) {
	b.Run("basic", func(b *testing.B) {
		var (
			body = ioutil.NopCloser(bytes.NewBuffer(nil))

			innerRes = &http.Response{
				StatusCode: 200,
				Body:       body,
			}

			inner = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				// pretend to be the Transport
				t := httptrace.ContextClientTrace(req.Context())
				t.WroteHeaders()
				t.WroteRequest(httptrace.WroteRequestInfo{})
				t.GotFirstResponseByte()
				return innerRes, nil
			})
		)

		b.ReportAllocs()

		r := RoundTrip(inner, nil)
		for i := 0; i < b.N; i++ {
			innerRes.Body = body

			res, err := r.RoundTrip(new(http.Request))
			if err != nil {
				panic(err)
			}
			res.Body.Close()
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			var (
				body = ioutil.NopCloser(bytes.NewBuffer(nil))

				innerRes = &http.Response{
					StatusCode: 200,
					Body:       body,
				}

				inner = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					// pretend to be the Transport
					t := httptrace.ContextClientTrace(req.Context())
					t.WroteHeaders()
					t.WroteRequest(httptrace.WroteRequestInfo{})
					t.GotFirstResponseByte()
					return innerRes, nil
				})
			)

			r := RoundTrip(inner, nil)
			for pb.Next() {
				innerRes.Body = body

				res, err := r.RoundTrip(new(http.Request))
				if err != nil {
					panic(err)
				}
				res.Body.Close()
			}
		})
	})

}
