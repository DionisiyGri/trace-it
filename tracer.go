package tracer

import (
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

// traceState record timestamps and track state in each hook to calculate durations
type traceState struct {
	Start            time.Time
	DNSStart         time.Time
	DNSDone          time.Time
	ConnectStart     time.Time
	ConnectDone      time.Time
	TLSStart         time.Time
	TLSDone          time.Time
	GotConn          time.Time
	FirstByte        time.Time
	Done             time.Time
	ConnectionIdle   time.Duration
	ConnectionReused bool
	RemoteAddr       string
}

// newTracer capture current timestamp in each hook
func newTracer(t *traceState) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			t.DNSStart = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			t.DNSDone = time.Now()
		},
		ConnectStart: func(_, _ string) {
			t.ConnectStart = time.Now()
		},
		ConnectDone: func(_, _ string, _ error) {
			t.ConnectDone = time.Now()
		},
		TLSHandshakeStart: func() {
			t.TLSStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			t.TLSDone = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			t.GotConn = time.Now()
			t.ConnectionReused = info.Reused
			t.ConnectionIdle = info.IdleTime

			if info.Conn != nil {
				t.RemoteAddr = info.Conn.RemoteAddr().String()
			}
		},
		GotFirstResponseByte: func() {
			t.FirstByte = time.Now()
		},
	}
}
