package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	server := http.Server{
		Addr: "127.0.0.1:8080",
	}
	http.HandleFunc("/convert", convert)

	http.Handle("/", http.FileServer(http.Dir("./static")))

	server.ListenAndServe()
}

func convert(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("subtitlefile")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	contents := string(data)
	name := header.Filename
	attach := fmt.Sprintf("attachment; filename=%s", name)
	// copy the relevant headers and filename:
	w.Header().Set("Content-Disposition", attach)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

	modtime := time.Now()
	http.ServeContent(w, r, name, modtime, strings.NewReader(contents))

	return
}
