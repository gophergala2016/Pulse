package router

import (
	"Pulse/pulse"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gophergala2016/Pulse/pulse/file"
)

// Result : is used for ResponseWriter in handlers
type Result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var buffStrings []string

// Start : will start the REST API
func Start() {
	http.HandleFunc("/log/message", StreamLog)
	http.HandleFunc("/log/file", SendFile)
	http.ListenAndServe(":8080", nil)
}

// StreamLog : Post log statement to our API
func StreamLog(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	result, _ := json.Marshal(Result{200, "success"})
	io.WriteString(w, string(result))

}

// SendFile : Post log files to our API
func SendFile(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	f, header, err := r.FormFile("file")

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	defer f.Close()

	stdIn := make(chan string)
	defer func() {
		dumpStringBuffer()
	}()
	pulse.Run(stdIn, addToBuffer)
	line := make(chan string)
	file.StreamRead(f, line)
	for l := range line {
		stdIn <- l
	}
	close(stdIn)

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)
}

func addToBuffer(value string) {
	buffStrings = append(buffStrings, value)
	if len(buffStrings) > 10 {
		dumpStringBuffer()
	}
}

func dumpStringBuffer() {
	file.Write(outputFile, buffStrings)
	buffStrings = nil
}
