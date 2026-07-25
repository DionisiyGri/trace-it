package waterpool

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/DionisiyGri/trace-it/tracer"
)

const TestURL = "https://tratt.net/laurie/blog/2026/test_case_reducers_are_underappreciated_debugging_tools.html"
const ReqPerClient = 20
const NumOfClients = 2

type Input struct {
	ID     int
	Client *http.Client
	URL    string
}

type Result struct {
	ClientID int
	Count    int

	Fastest time.Duration
	Slowest time.Duration
	Avg     time.Duration

	Err error
}

type ResponseStats struct {
	Results   []Result
	GlobalAvg time.Duration
}

func Do(input Input) {
	resp, err := input.Client.Get(input.URL) // TODO: add context cancelations or smth
	if err != nil {
		log.Println(err)
	}

	//close response body to reuse connections across clients
	_, err = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

// TODO create client withlog/withresult/withsmth
func CreateClient(id int, agg *Aggregator) *http.Client {
	return &http.Client{
		Transport: &tracer.TracingTransport{
			OnResult: func(tr tracer.TraceResult) {
				agg.Add(id, tr)
			},
			Log: func(req *http.Request, t *tracer.Timings) {
				log.Printf("[%s %s] || dns_lookup=%v | tls_handshake=%v | server_response=%v | total=%v",
					req.Method, req.URL,
					t.DNSDone.Sub(t.DNSStart),
					t.TLSDone.Sub(t.TLSStart),
					t.FirstByte.Sub(t.GotConn),
					t.DNSDone.Sub(t.Start),
				)
			},
		},
	}
}

func ShowStats(agg *Aggregator) ResponseStats {
	var (
		globalCount int
		globalTotal time.Duration
		response    ResponseStats
	)

	for clientID, s := range agg.Clients {
		avg := s.Total / time.Duration(s.Count)
		response.Results = append(response.Results, Result{
			ClientID: clientID,
			Count:    s.Count,
			Fastest:  s.Fastest,
			Slowest:  s.Slowest,
			Avg:      avg,
		})

		globalCount += s.Count
		globalTotal += s.Total
	}

	if globalCount > 0 {
		response.GlobalAvg = globalTotal / time.Duration(globalCount)
	}
	return response
}

func printStats(agg *Aggregator) {
	var (
		globalCount int
		globalTotal time.Duration
	)

	for clientID, s := range agg.Clients {
		avg := s.Total / time.Duration(s.Count)

		log.Printf(
			"client=%d requests=%d avg=%v fastest=%v slowest=%v",
			clientID, s.Count, avg, s.Fastest, s.Slowest,
		)

		globalCount += s.Count
		globalTotal += s.Total
	}

	if globalCount > 0 {
		log.Printf("global avg=%v", globalTotal/time.Duration(globalCount))
	}
}
