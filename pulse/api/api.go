package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Result : is used for ResponseWriter in handlers
type Result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

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

	file, header, err := r.FormFile("file")

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "bad request"})
		io.WriteString(w, string(result))
		return
	}

	defer file.Close()

	out, err := os.Create("/tmp/uploadedfile")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		result, _ := json.Marshal(Result{400, "Unable to create the file for writing."})
		io.WriteString(w, string(result))
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	fmt.Fprintf(w, "File uploaded successfully : ")
	fmt.Fprintf(w, header.Filename)
}
