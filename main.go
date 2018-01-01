package main

import (
	"bytes"
	"fmt"
	"io"
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
	var Buf bytes.Buffer
	modtime := time.Now()

	file, header, err := r.FormFile("subtitlefile")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	name := strings.Split(header.Filename, ".")[0]

	// Copy the file data to my buffer
	io.Copy(&Buf, file)
	// do something with the contents...
	// I normally have a struct defined and unmarshal into a struct, but this will
	// work as an example

	contents := Buf.String()

	// I reset the buffer in case I want to use it again
	// reduces memory allocations in more intense projects
	Buf.Reset()
	// do something else:
	// copy the relevant headers. If you want to preserve the downloaded file name, extract it with go's url parser.
	attach := fmt.Sprintf("attachment; filename=%s", header.Filename)
	w.Header().Set("Content-Disposition", attach)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

	http.ServeContent(w, r, name, modtime, strings.NewReader(contents))

	return
}
