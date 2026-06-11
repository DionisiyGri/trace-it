package main

import (
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

// Wrapper around default transport with tracer and logger
type TracingTransport struct {
	RoundTripper http.RoundTripper
	Log          func(req *http.Request, t *timings)
	Result       TraceResult
}

type TraceResult struct {
	DNSLookup        string
	TLSHandshake     string
	Server1bResponse string
	Total            string
}

func (tt *TracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := tt.RoundTripper
	if transport == nil {
		transport = http.DefaultTransport
	}

	t := &timings{start: time.Now()}
	trace := newTracer(t)
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	res, err := transport.RoundTrip(req) // RoundTrip returns when the response headers are read, not when the body is consumed
	res.Body = &timedBody{
		ReadCloser: res.Body,
		onClose:    func() { t.done = time.Now() }, // helps to calculate total request duration (headers+body)
	}

	t.done = time.Now()
	if tt.Log != nil {
		tt.Log(req, t)
	}

	tt.Result.DNSLookup = t.dnsDone.Sub(t.dnsStart).String()
	tt.Result.TLSHandshake = t.tlsDone.Sub(t.tlsStart).String()
	tt.Result.Server1bResponse = t.firstByte.Sub(t.gotConn).String()
	tt.Result.Total = t.done.Sub(t.start).String()

	return res, err
}

func (tt *TracingTransport) GetTraceResults() TraceResult {
	return tt.Result
}

type timedBody struct {
	io.ReadCloser
	onClose func()
}

func (tb *timedBody) Close() error {
	tb.onClose()
	return tb.ReadCloser.Close()
}

var _ http.RoundTripper = (*TracingTransport)(nil)
