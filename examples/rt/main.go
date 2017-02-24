package main

import (
	"context"
	"log"
	"net/http"

	"github.com/rakyll/trace"
	"github.com/rakyll/trace/gcp"
)

type tracerRT struct {
	base http.RoundTripper
}

func (rt *tracerRT) RoundTrip(req *http.Request) (*http.Response, error) {
	_, finish := trace.ChildSpan(req.Context(), req.URL.Host+req.URL.Path)
	defer finish()

	return rt.base.RoundTrip(req)
}

func main() {
	tc, err := gcp.NewClient(context.Background(), "bamboo-shift-504")
	if err != nil {
		log.Fatal(err)
	}

	client := http.Client{Transport: &tracerRT{base: http.DefaultTransport}}
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("GET", "https://google.com", nil)
		ctx := trace.WithClient(req.Context(), tc)
		req = req.WithContext(ctx) // enable tracing for this request

		if _, err := client.Do(req); err != nil {
			log.Println(err)
		}
	}
}
