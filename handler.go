package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

func convert(w http.ResponseWriter, r *http.Request) {
	// This limits the size of the entire request body and not
	// an individual file. Since we are uploading a single file
	// at a time, limiting the size of the request body should
	// a good approximation of limiting the file size to 200kB:
	r.Body = http.MaxBytesReader(w, r.Body, 200*1024)
	file, header, err := r.FormFile("subtitlefile")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	seconds, err := strconv.ParseFloat(r.FormValue("seconds"), 64)
	if err != nil {
		log.Fatal(err)
	}

	plusmin, err := strconv.ParseFloat(r.FormValue("plusmin"), 64)
	if err != nil {
		log.Fatal(err)
	}
	seconds *= plusmin

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	contents := string(data)

	nameIn := header.Filename

	fromExt := path.Ext(nameIn)
	if !allowedExt(fromExt) {
		log.Fatalf("'From' extension '%s' is not allowed.", fromExt)
	}
	toExt := r.FormValue("to")
	if !allowedExt(toExt) {
		log.Fatalf("'To' extension '%s' is not allowed.", toExt)
	}

	nameOut := nameOutput(nameIn, seconds, fromExt, toExt)

	// We need vtt to convert, because vtt has '.' decimals
	// (instead of ',' decimals in srt)
	if fromExt == ".srt" {
		contents = toVTT(contents)
	}

	contents = convertVTT(contents, seconds)

	if toExt == ".srt" {
		contents = toSRT(contents)
	}

	attach := fmt.Sprintf("attachment; filename=%s", nameOut)
	// copy the relevant headers and filename:
	w.Header().Set("Content-Disposition", attach)
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

	modtime := time.Now()
	http.ServeContent(w, r, nameOut, modtime, strings.NewReader(contents))

	return
}
