package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type health struct {
	OK      bool   `json:"ok"`
	Service string `json:"service"`
	Port    string `json:"port"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3400"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(health{OK: true, Service: "go-basic", Port: port})
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintf(w, "go-basic running on %s\n", port)
	})

	addr := ":" + port
	log.Printf("[go-basic] listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}
