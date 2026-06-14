package main

import (
	"sync"
	"time"

	"github.com/DionisiyGri/trace-it/tracer"
)

type Aggregator struct {
	mu sync.Mutex

	Clients map[int]*ClientStats
}

type ClientStats struct {
	Count   int
	Total   time.Duration
	Fastest time.Duration
	Slowest time.Duration
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		Clients: make(map[int]*ClientStats),
	}
}

func (a *Aggregator) Add(clientID int, tr tracer.TraceResult) {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.Clients[clientID]
	if !ok {
		s = &ClientStats{}
		a.Clients[clientID] = s
	}

	s.Count++
	s.Total += tr.Total

	if s.Fastest == 0 || tr.Total < s.Fastest {
		s.Fastest = tr.Total
	}

	if tr.Total > s.Slowest {
		s.Slowest = tr.Total
	}
}
