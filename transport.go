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
	OnResult     func(TraceResult)
}

// TraceResult contains metadata and timing information collected for a completed HTTP request
type TraceResult struct {
	URL            string
	Method         string
	StatusCode     int
	Timings        Timings
	ConnectionInfo ConnectionInfo
}

// Timings contains the measured durations of each request phase
type Timings struct {
	DNSLookup        time.Duration
	TLSHandshake     time.Duration
	Server1bResponse time.Duration
	Total            time.Duration
}

// ConnectionInfo contains extra info about connection
type ConnectionInfo struct {
	IsReused   bool
	Idle       time.Duration
	RemoteAddr string
}

// wrapper around original body with exta info
type timedBody struct {
	io.ReadCloser

	traceState *traceState

	onResult func(TraceResult)
}

func (tt *TracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := tt.RoundTripper
	if transport == nil {
		transport = http.DefaultTransport
	}

	t := &traceState{Start: time.Now()}
	trace := newTracer(t)

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	res, err := transport.RoundTrip(req) // NB: RoundTrip returns when the response headers are read, not when the body is consumed
	if res != nil {
		res.Body = &timedBody{
			ReadCloser: res.Body,
			traceState: t,
			onResult: func(r TraceResult) {
				r.URL = req.URL.String()
				r.Method = req.Method
				r.StatusCode = res.StatusCode

				if tt.OnResult != nil {
					tt.OnResult(r)
				}
			},
		}
	}

	return res, err
}

func (tb *timedBody) Close() error {
	tb.traceState.Done = time.Now()

	if tb.onResult != nil {
		tb.onResult(TraceResult{
			Timings: Timings{
				DNSLookup:        tb.traceState.DNSDone.Sub(tb.traceState.DNSStart),
				TLSHandshake:     tb.traceState.TLSDone.Sub(tb.traceState.TLSStart),
				Server1bResponse: tb.traceState.FirstByte.Sub(tb.traceState.GotConn),
				Total:            tb.traceState.Done.Sub(tb.traceState.Start),
			},
			ConnectionInfo: ConnectionInfo{
				IsReused:   tb.traceState.ConnectionReused,
				Idle:       tb.traceState.ConnectionIdle,
				RemoteAddr: tb.traceState.RemoteAddr,
			},
		})
	}

	return tb.ReadCloser.Close()
}

var _ http.RoundTripper = (*TracingTransport)(nil)
