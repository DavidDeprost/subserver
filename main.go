package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static")))
	mux.HandleFunc("/convert", convert)

	server := http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	log.Printf("\nSubserver is now running on http://%s ...\n", server.Addr)
	server.ListenAndServe()
}
