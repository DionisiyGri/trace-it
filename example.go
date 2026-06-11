package main

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := template.Must(template.ParseFiles("./result.html")).Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func traceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, _ := traceRequest(string(body))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func traceRequest(url string) (*TraceResult, error) {
	var transport TracingTransport

	client := &http.Client{
		Transport: &transport,
	}

	_, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	return &transport.Result, nil
}
