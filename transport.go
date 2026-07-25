package tracer

import (
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

// Wrapper around default transport with tracer and logger
type TracingTransport struct {
	RoundTripper http.RoundTripper
	Log          func(req *http.Request, t *Timings)
	OnResult     func(TraceResult)
	Result       TraceResult
}

type TraceResult struct {
	DNSLookup        time.Duration
	TLSHandshake     time.Duration
	Server1bResponse time.Duration
	Total            time.Duration
}

func (tt *TracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := tt.RoundTripper
	if transport == nil {
		transport = http.DefaultTransport
	}

	t := &Timings{Start: time.Now()}
	trace := newTracer(t)
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	res, err := transport.RoundTrip(req) // RoundTrip returns when the response headers are read, not when the body is consumed
	if res != nil {
		res.Body = &timedBody{
			ReadCloser: res.Body,
			onResult:   tt.OnResult,
			timings:    t,
		}
	}

	if tt.Log != nil {
		tt.Log(req, t)
	}

	return res, err
}

type timedBody struct {
	io.ReadCloser
	timings *Timings

	onClose  func()
	onResult func(TraceResult)
}

func (tb *timedBody) Close() error {
	tb.timings.Done = time.Now()

	if tb.onResult != nil {
		tb.onResult(TraceResult{
			DNSLookup:        tb.timings.DNSDone.Sub(tb.timings.DNSStart),
			TLSHandshake:     tb.timings.TLSDone.Sub(tb.timings.TLSStart),
			Server1bResponse: tb.timings.FirstByte.Sub(tb.timings.GotConn),
			Total:            tb.timings.Done.Sub(tb.timings.Start),
		})
	}
	return tb.ReadCloser.Close()
}

var _ http.RoundTripper = (*TracingTransport)(nil)
