package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	host := os.Getenv("HOST_TRACEIT")
	if host == "" {
		log.Fatalf("host is not provided")
	}

	port := os.Getenv("PORT_TRACEIT")
	if port == "" {
		log.Fatalf("port is nor provided")
	}

	addr := fmt.Sprintf("[%s]:%s", host, port)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/trace", traceHandler)

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}
