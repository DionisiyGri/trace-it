package main

import (
	"net/http"
)

func main() {
	// url := "https://www.lucasfcosta.com/blog/backpressure-is-all-you-need"

	// client := &http.Client{
	// 	Transport: &TracingTransport{
	// 		Log: func(req *http.Request, t *timings) {
	// 			log.Printf("%s %s >>> dns_lookup=%v tls_handshake=%v server_1b_response=%v total=%v",
	// 				req.Method, req.URL,
	// 				t.dnsDone.Sub(t.dnsStart),
	// 				t.tlsDone.Sub(t.tlsStart),
	// 				t.firstByte.Sub(t.gotConn),
	// 				t.done.Sub(t.start),
	// 			)
	// 		},
	// 	},
	// }
	// client.Get(url)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/trace", traceHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
