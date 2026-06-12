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

	t := &Timings{Start: time.Now()}
	trace := newTracer(t)
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	res, err := transport.RoundTrip(req) // RoundTrip returns when the response headers are read, not when the body is consumed
	res.Body = &timedBody{
		ReadCloser: res.Body,
		onClose:    func() { t.Done = time.Now() }, // helps to calculate total request duration (headers+body)
	}

	t.Done = time.Now()
	if tt.Log != nil {
		tt.Log(req, t)
	}

	tt.Result.DNSLookup = t.DNSDone.Sub(t.DNSStart).String()
	tt.Result.TLSHandshake = t.TLSDone.Sub(t.TLSStart).String()
	tt.Result.Server1bResponse = t.FirstByte.Sub(t.GotConn).String()
	tt.Result.Total = t.Done.Sub(t.Start).String()

	return res, err
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
