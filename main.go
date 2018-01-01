package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	file, header, err := r.FormFile("subtitlefile")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	name := strings.Split(header.Filename, ".")
	fmt.Fprintf(w, "File name: %s\n\n", name[0])

	// Copy the file data to my buffer
	io.Copy(&Buf, file)
	// do something with the contents...
	// I normally have a struct defined and unmarshal into a struct, but this will
	// work as an example

	contents := Buf.String()
	fmt.Fprintln(w, contents)
	// I reset the buffer in case I want to use it again
	// reduces memory allocations in more intense projects
	Buf.Reset()
	// do something else
	// etc write header
	return
}
