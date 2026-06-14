package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/DionisiyGri/trace-it/examples/waterpool/waterpool"
)

//go:embed templates
var tplFolder embed.FS

type TraceInput struct {
	URL         string `json:"url"`
	NumClients  int    `json:"clients"`
	NumRequests int    `json:"calls"`
}

type TraceResponse struct {
	Clients   []ClientStats `json:"clients"`
	GlobalAvg string        `json:"global_avg"`
}

type ClientStats struct {
	ClientID int `json:"client"`
	Count    int `json:"num_req"`

	Fastest string `json:"fastest"`
	Slowest string `json:"slowest"`
	Avg     string `json:"avg"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(tplFolder, "templates/trace-it.html")
	if err != nil {
		log.Fatal(err)
	}

	tmpl.Execute(w, nil)
}

func traceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var input TraceInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isUrl(input.URL) {
		http.Error(w, "Enter valid URL, suka", http.StatusBadRequest)
		return
	}

	res, err := traceRequest(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func traceRequest(input TraceInput) (TraceResponse, error) {
	agg := waterpool.NewAggregator()

	var wg sync.WaitGroup
	for i := 0; i < input.NumClients; i++ {
		cl := waterpool.CreateClient(i, agg)

		wg.Add(1)
		go func(id int, client *http.Client) {
			defer wg.Done()
			for j := 0; j < input.NumRequests; j++ {
				waterpool.Do(waterpool.Input{ID: i, Client: cl, URL: input.URL})
			}
		}(i, cl)
	}
	wg.Wait()

	stats := waterpool.ShowStats(agg)

	resp := TraceResponse{
		Clients:   make([]ClientStats, 0, len(stats.Results)),
		GlobalAvg: stats.GlobalAvg.String(),
	}

	for _, st := range stats.Results {
		resp.Clients = append(resp.Clients, ClientStats{
			ClientID: st.ClientID,
			Count:    st.Count,
			Fastest:  st.Fastest.String(),
			Slowest:  st.Slowest.String(),
			Avg:      st.Avg.String(),
		})
	}
	return resp, nil
}

func isUrl(str string) bool {
	url, err := url.ParseRequestURI(str)
	if err != nil {
		log.Print(err.Error())
		return false
	}

	address := net.ParseIP(url.Host)

	if address == nil {
		log.Print("nil address; host=", url.Host)
		return strings.Contains(url.Host, ".")
	}

	return true
}
