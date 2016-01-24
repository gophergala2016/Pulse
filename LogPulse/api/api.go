// Package api is to start the api to be listening on different endpoints
// The API will listen on the port specified in the PulseConfig.toml.
// There are 2 endpoints:
// POST /log/file this will read the file line by line passing in each line to the algorithm
// POST /log/message (in development) this will take a string and pass it directly to the algorithm
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/gophergala2016/Pulse/LogPulse/config"
	"github.com/gophergala2016/Pulse/LogPulse/email"
	"github.com/gophergala2016/Pulse/LogPulse/file"
	"github.com/gophergala2016/Pulse/pulse"
)

// Result is used for ResponseWriter in handlers
type Result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var buffStrings []string
var port int

func init() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			os.Exit(0)
		}
	}()
	val, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("API: %s", err))
	}
	port = val.Port
}

// Start will run the REST API.
func Start() {
	http.HandleFunc("/log/message", StreamLog)
	http.HandleFunc("/log/file", SendFile)

	fmt.Printf("Listening on localhost:%d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}

// StreamLog listens for post request for a string value.
// This string is then passed to the algorithm for analyzing.
func StreamLog(w http.ResponseWriter, r *http.Request) {

	// Checking to see if the request was a post.
	// If not return a 400: bad request
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	// Decoding the body of the response.
	// If we could not parse it as json then respond with a 400: bad request
	decoder := json.NewDecoder(r.Body)
	var body struct {
		Message string `json:"message"`
	}
	err := decoder.Decode(&body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	//TODO: post the string to the algorthm for analyzing

	// If we were able to decode and send string to algorithm return a 200: success
	w.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(Result{200, "success"})
	io.WriteString(w, string(result))

}

// SendFile listens for a POST that has a form field named file and email in the body.
// Using the file field we will download the specified file to the server.
// The email field is used to email the user the results once algorithm is done.
func SendFile(w http.ResponseWriter, r *http.Request) {

	// Checking to see if the request was a post.
	// If not return a 400: bad request
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}
	compressed := false
	// Get the file field from the form in the response.
	// If we cannot parse it a 400 bad request is returned
	f, header, err := r.FormFile("file")
	fmt.Println("Form File")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	defer f.Close()

	var body struct {
		Email string `json:"email"`
	}
	// Parse the Form in the response.
	// Check if the email field is a valid email.
	// If not return an 400: bad request
	r.ParseForm()
	body.Email = r.Form["email"][0]
	if !email.IsValid(body.Email) {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "email not valid"})
		io.WriteString(w, string(result))
		return
	}
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}
	fmt.Println("Received bodys", body.Email)
	extension := filepath.Ext(header.Filename)
	filename := header.Filename[0 : len(header.Filename)-len(extension)]

	if extension == ".gz" {
		// Load compressed file on disk
		out, err := os.Create(fmt.Sprintf("%s.gz", filename))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			result, _ := json.Marshal(Result{400, "bad request"})
			io.WriteString(w, string(result))
			return
		}

		defer out.Close()

		// Write the content from POST to the file
		_, err = io.Copy(out, f)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			result, _ := json.Marshal(Result{400, "gzip copy failed"})
			io.WriteString(w, string(result))
			return
		}

		// Uncompress file
		err = file.UnGZip(fmt.Sprintf("%s.gz", filename))
		if err != nil {
			spew.Dump(err)
			w.Header().Set("Content-Type", "application/json")
			result, _ := json.Marshal(Result{400, "gzip uncompressed failed"})
			io.WriteString(w, string(result))
			return
		}

		compressed = true
	}

	stdIn := make(chan string)
	email.ByPassMail = true // Needs to bypass emails and store in JSON
	email.OutputFile = filename + ".json"
	email.EmailList = []string{body.Email}

	// Run on separat go routine so that we can give users a response on page first.
	go func() {
		start := time.Now()
		// Start the pulse algorithm
		pulse.Run(stdIn, email.SaveToCache)
		line := make(chan string)

		if compressed {
			file.Read(filename, line)
		} else {
			file.StreamRead(f, line)
		}

		for l := range line {
			if l == "EOF" {
				email.ByPassMail = false
				// Once EOF, time to send email from cache JSON storage
				email.DumpBuffer() // Must clear buffer first before sending
				email.SendFromCache(email.OutputFile)
				break
			}
			stdIn <- l
		}
		close(stdIn)

		elapsed := time.Since(start)
		log.Printf("Pulse Algorithm took %s", elapsed)

		// Clean up
		defer func() {
			if compressed {
				err := os.Remove(filename)
				if err != nil {
					fmt.Println("Failed to delete uncompressed file, please delete")
				}

				err = os.Remove(fmt.Sprintf("%s.gz", filename))
				if err != nil {
					fmt.Println("Failed to delete uncompressed file, please delete")
				}
			}
		}()
	}()

	// Return a 200 success even if algorithm is still going.
	w.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(Result{200, "success"})
	io.WriteString(w, string(result))
}
