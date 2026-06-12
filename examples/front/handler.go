package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/DionisiyGri/trace-it/tracer"
)

//go:embed templates
var tplFolder embed.FS

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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input := string(body)
	if !IsUrl(input) {
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

func traceRequest(url string) (*tracer.TraceResult, error) {
	var transport tracer.TracingTransport

	client := &http.Client{
		Transport: &transport,
	}

	_, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	return &transport.Result, nil
}

func IsUrl(str string) bool {
	url, err := url.ParseRequestURI(str)
	if err != nil {
		log.Print(err.Error())
		return false
	}

	address := net.ParseIP(url.Host)
	log.Print("host=", address)

	if address == nil {
		log.Print("nil address; host=", url.Host)
		return strings.Contains(url.Host, ".")
	}

	return true
}
