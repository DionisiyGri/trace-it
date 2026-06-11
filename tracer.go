package main

import (
	"crypto/tls"
	"log"
	"net/http/httptrace"
	"time"
)

// timings record timestamps in each hook to calculate durations
type timings struct {
	start        time.Time
	dnsStart     time.Time
	dnsDone      time.Time
	connectStart time.Time
	connectDone  time.Time
	tlsStart     time.Time
	tlsDone      time.Time
	gotConn      time.Time
	firstByte    time.Time
	done         time.Time
}

// newTracer capture current timestamp in each hook
func newTracer(t *timings) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			t.dnsStart = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			t.dnsDone = time.Now()
		},
		ConnectStart: func(_, _ string) {
			t.connectStart = time.Now()
		},
		ConnectDone: func(_, _ string, _ error) {
			t.connectDone = time.Now()
		},
		TLSHandshakeStart: func() {
			t.tlsStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			t.tlsDone = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			t.gotConn = time.Now()
			if info.Reused {
				log.Printf("connection reused (idle for %v)", info.IdleTime)
			} else {
				log.Printf("new connection to %s", info.Conn.RemoteAddr())
			}
		},
		GotFirstResponseByte: func() {
			t.firstByte = time.Now()
		},
	}
}
