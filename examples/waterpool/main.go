// package fan-in has a main purpose to test a scenario when multiple clients hit single URL
package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/DionisiyGri/trace-it/examples/waterpool/waterpool"
	"github.com/DionisiyGri/trace-it/tracer"
)

type Result struct {
	ClientID int
	Err      error
	Trace    tracer.TraceResult
}

func main() {
	log.Print("start")

	var wg sync.WaitGroup
	agg := waterpool.NewAggregator()

	for i := 0; i < waterpool.NumOfClients; i++ {
		cl := waterpool.CreateClient(i, agg)

		wg.Add(1)
		go func(id int, client *http.Client) {
			defer wg.Done()
			for j := 0; j < waterpool.ReqPerClient; j++ {
				waterpool.Do(waterpool.Input{ID: i, Client: cl, URL: waterpool.TestURL})
			}
		}(i, cl)
	}

	wg.Wait()

	log.Print(waterpool.ShowStats(agg))
	log.Print("stop")
}
