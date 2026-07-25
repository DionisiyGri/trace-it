package tracer

import (
	"crypto/tls"
	"log"
	"net/http/httptrace"
	"time"
)

// Timings record timestamps in each hook to calculate durations
type Timings struct {
	Start        time.Time
	DNSStart     time.Time
	DNSDone      time.Time
	ConnectStart time.Time
	ConnectDone  time.Time
	TLSStart     time.Time
	TLSDone      time.Time
	GotConn      time.Time
	FirstByte    time.Time
	Done         time.Time
}

// newTracer capture current timestamp in each hook
func newTracer(t *Timings) *httptrace.ClientTrace {
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
			if info.Reused {
				log.Printf("connection reused (idle for %v)", info.IdleTime)
			} else {
				log.Printf("new connection to %s", info.Conn.RemoteAddr())
			}
		},
		GotFirstResponseByte: func() {
			t.FirstByte = time.Now()
		},
	}
}
