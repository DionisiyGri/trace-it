// package fan-in has a main purpose to test a scenario when multiple clients hit single URL
package main

import (
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/DionisiyGri/trace-it/tracer"
)

const _testURL = "https://tratt.net/laurie/blog/2026/test_case_reducers_are_underappreciated_debugging_tools.html"
const _reqPerClient = 20
const _numOfClients = 2

type Result struct {
	ClientID int
	Err      error
	Trace    tracer.TraceResult
}

func main() {
	log.Print("start")

	var wg sync.WaitGroup
	agg := NewAggregator()

	for i := 0; i < _numOfClients; i++ {
		cl := createClient(i, agg)

		wg.Add(1)
		go func(id int, client *http.Client) {
			defer wg.Done()
			for j := 0; j < _reqPerClient; j++ {
				req(i, cl)
			}
		}(i, cl)
	}

	wg.Wait()
	printStats(agg)

	log.Print("stop")
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

func createClient(id int, agg *Aggregator) *http.Client {
	return &http.Client{
		Transport: &tracer.TracingTransport{
			OnResult: func(tr tracer.TraceResult) {
				agg.Add(id, tr)
			},
		},
	}
}

func req(id int, cl *http.Client) Result {
	resp, err := cl.Get(_testURL) // TODO: add context cancelations or smth
	if err != nil {
		log.Println(err)
	}

	//close response body to reuse connections across clients
	_, err = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return Result{
		Err:      err,
		ClientID: id,
	}
}
