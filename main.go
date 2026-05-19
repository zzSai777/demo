package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"go-demo/internal/control"
)

func newServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Go demo!")
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.Handle("/control/v1/", control.NewHandler(control.NewMemoryStore()))
	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("go-demo listening on %s", addr)
	if err := http.ListenAndServe(addr, newServer()); err != nil {
		log.Fatal(err)
	}
}
