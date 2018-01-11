package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static")))
	mux.HandleFunc("/convert", convert)

	server := http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	log.Printf("Subserver is now running on http://%s ...\n", server.Addr)
	server.ListenAndServe()
}

func allowedExt(ext string) bool {
	slice := []string{".srt", ".vtt"}

	for _, val := range slice {
		if val == ext {
			return true
		}
	}
	return false
}

func convert(w http.ResponseWriter, r *http.Request) {
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

	fromExt := r.FormValue("from")
	if !allowedExt(fromExt) {
		log.Fatalf("'From' extension '%s' is not allowed.", fromExt)
	}
	toExt := r.FormValue("to")
	if !allowedExt(toExt) {
		log.Fatalf("'To' extension '%s' is not allowed.", toExt)
	}
	nameIn := header.Filename
	nameOut := nameOutput(nameIn, seconds, fromExt, toExt)

	if fromExt != path.Ext(nameIn) {
		fmt.Printf("from = %s\n", fromExt)
		fmt.Printf("ext = %s\n", path.Ext(nameIn))
		log.Fatal("The chosen 'from' extension does not match ",
			"that of the filename.")
	}

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

// Determines the name of the outputfile based on the inputfile and seconds;
// the name of the new file is identical to the old one, but prepended with "{+x.xx_Sec}_".
//
// However, if the file has already been processed by submod before, we simply change
// the 'increment number' x, instead of prepending "{+x.xx_Sec}_" a second time.
// This way we can conveniently process files multiple times, and still have sensible names.
func nameOutput(inputfile string, seconds float64, fromExt string, toExt string) string {
	// Regex to check if the inputfile was previously processed by submod:
	proc, err := regexp.Compile(`\{[+-]\d+\.\d+_Sec\}_`)
	if err != nil {
		log.Fatal(err)
	}

	var processed bool = proc.MatchString(inputfile)
	var placeholder string
	var incr float64

	// Inputfile was processed by submod previously:
	if processed {
		// Regex for extracting the increment number from the inputfile:
		re, err := regexp.Compile(`[+-]\d+\.\d+`)
		if err != nil {
			log.Fatal(err)
		}

		// FindString extracts the leftmost occurrence of 're'
		var number string = re.FindString(inputfile)

		incr, err = strconv.ParseFloat(number, 64)
		if err != nil {
			log.Fatal("\nError processing seconds for filename:\n", err)
		}
		incr += seconds

		// Apparently golang does not have a single replace regex method,
		// so we have to get creative; FindStringIndex returns the start
		// to end indices of the leftmost occurrence of proc as a slice,
		// which we then use to replace proc with the format:
		index := proc.FindStringIndex(inputfile)
		placeholder = "{%.2f_Sec}_" + inputfile[index[1]:]
	} else {
		incr = seconds
		placeholder = "{%.2f_Sec}_" + inputfile
	}

	if incr >= 0 {
		placeholder = "{+" + placeholder[1:]
	}

	var outputfile string = fmt.Sprintf(placeholder, incr)

	if fromExt != toExt {
		outputfile = strings.TrimSuffix(outputfile, fromExt) + toExt
	}

	return outputfile
}

// Converts an SRT subtitle file to a VTT.
func toVTT(contents string) string {
	// Compile regex to find time-line outside of loop for performance:
	re, err := regexp.Compile(`\d\d:\d\d:\d\d\,\d\d\d`)
	if err != nil {
		log.Fatal(err)
	}

	var buffer bytes.Buffer

	// Iterate line by line over contents:
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {

		var old_line string = scanner.Text()
		var new_line string
		var time_line bool = re.MatchString(old_line)

		// Time-line: This is the line we need to modify
		if time_line {
			// We need '.' instead of ',' for VTT (and floats)!
			new_line = strings.Replace(old_line, ",", ".", 2)
		} else {
			new_line = old_line
		}

		buffer.WriteString(new_line + "\n")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return buffer.String()
}

// Converts a VTT subtitle file to an SRT.
func toSRT(contents string) string {
	// Compile regex to find time-line outside of loop for performance:
	re, err := regexp.Compile(`\d\d:\d\d:\d\d\.\d\d\d`)
	if err != nil {
		log.Fatal(err)
	}

	var buffer bytes.Buffer

	// Iterate line by line over contents:
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {

		var old_line string = scanner.Text()
		var new_line string
		var time_line bool = re.MatchString(old_line)

		// Time-line: This is the line we need to modify
		if time_line {
			// We need ',' instead of '.' for SRT files!
			new_line = strings.Replace(old_line, ".", ",", 2)
		} else {
			new_line = old_line
		}

		buffer.WriteString(new_line + "\n")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return buffer.String()
}

// Loops through the given inputfile, modifies the lines consisting of the time encoding,
// writes everything back to outputfile, and returns the number of subtitles that were deleted.
//
// This function is identical to convertSRT,
// except that it uses '.' for the seconds field's decimal space.
//
// The subtitle files consist of a repetition of the following 3 lines:
//
// - Index-line: integer count indicating line number
// - Time-line: encoding the duration for which the subtitle appears
// - Sub-line: the actual subtitle to appear on-screen (1 or 2 lines)
//
// Example .vtt (Note: '.' for decimal spaces):
//
// 1
// 00:00:00.243 --> 00:00:02.110
// Previously on ...
//
// 2
// 00:00:03.802 --> 00:00:05.314
// Etc.
func convertVTT(contents string, seconds float64) string {
	// Compile regex to find time-line outside of loop for performance:
	re, err := regexp.Compile(`\d\d:\d\d:\d\d\.\d\d\d`)
	if err != nil {
		log.Fatal(err)
	}

	var skip bool = false
	var buffer bytes.Buffer

	// Iterate line by line over contents:
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {

		var old_line string = scanner.Text()
		var new_line string
		var time_line bool = re.MatchString(old_line)

		// Time-line: This is the line we need to modify
		if time_line {
			new_line = processLine(old_line, seconds)
			if new_line == "(DELETED)\n" {
				skip = true
			}
		} else {
			// When skip = True, subtitles are shifted too far back
			// into the past (before the start of the movie),
			// so they are deleted:
			if skip {
				// Subtitles can be 1 or 2 lines; we should only update
				// skip when we have arrived at an empty line:
				if old_line == "" {
					skip = false
				}
				continue
			} else {
				new_line = old_line
			}
		}

		buffer.WriteString(new_line + "\n")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return buffer.String()
}

// Process the given line by adding seconds to start and end time.
// (subtracting if seconds is negative)
//
// Example line:  '00:00:01.913 --> 00:00:04.328'
// Index:          01234567890123456789012345678
// Index by tens: (0)        10        20     (28)
func processLine(line string, seconds float64) string {
	var start string = line[0:12]
	start = processTime(start, seconds)

	var end string = line[17:29]
	end = processTime(end, seconds)

	if start == "(DELETED)\n" {
		if end == "(DELETED)\n" {
			line = "(DELETED)\n"
		} else {
			line = "00:00:00.000 --> " + end
		}
	} else {
		line = start + " --> " + end
	}

	return line
}

// Increment the given time_string by 'incr' seconds
//
// The time-string has the form '00:00:00.000',
// and converts to the following format string:
// "%02d:%02d:%06.3f"
func processTime(time_string string, incr float64) string {
	hrs, err := strconv.Atoi(time_string[0:2])
	if err != nil {
		log.Fatal("\nError processing hours:\n", err)
	}
	mins, err := strconv.Atoi(time_string[3:5])
	if err != nil {
		log.Fatal("\nError processing minutes:\n", err)
	}
	secs, err := strconv.ParseFloat(time_string[6:12], 64)
	if err != nil {
		log.Fatal("\nError processing seconds:\n", err)
	}

	var hr time.Duration = time.Duration(hrs) * time.Hour
	var min time.Duration = time.Duration(mins) * time.Minute
	var sec time.Duration = time.Duration(secs*1000) * time.Millisecond
	var delta time.Duration = time.Duration(incr*1000) * time.Millisecond
	var new_time time.Duration = hr + min + sec + delta

	// incr can be negative, so the new time could be too:
	if new_time >= 0 {
		// NOT casting to int64 might be problematic on 32 bit systems though:
		// when int is 32 bits wide, it can't hold the largest of time.Duration values (which are 64 bit)!
		// But this shouldn't be a problem for the small values we expect.
		hrs = int(new_time / time.Hour)
		mins = int((new_time % time.Hour) / time.Minute)
		secs = float64((new_time%time.Minute)/time.Millisecond) / 1000
		time_string = fmt.Sprintf("%02d:%02d:%06.3f", hrs, mins, secs)
	} else {
		// new_time < 0: the subtitles are now scheduled before the start
		// of the movie, so we can delete them:
		time_string = "(DELETED)\n"
	}

	return time_string
}
