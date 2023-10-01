package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	http.HandleFunc("/transcript", transcriptHandler)
	http.ListenAndServe(":8080", nil)
}

func transcriptHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are supported", http.StatusMethodNotAllowed)
		return
	}

	// Parse the MP3 file from the request
	mp3File, _, err := r.FormFile("mp3_file")
	if err != nil {
		http.Error(w, "Failed to parse MP3 file", http.StatusBadRequest)
		return
	}
	defer mp3File.Close()

	// Create a temporary file to store the uploaded MP3
	tmpFile, err := ioutil.TempFile("", "uploaded-*.mp3")
	if err != nil {
		http.Error(w, "Failed to create temporary file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy the MP3 data to the temporary file
	_, err = io.Copy(tmpFile, mp3File)
	if err != nil {
		http.Error(w, "Failed to copy MP3 data", http.StatusInternalServerError)
		return
	}

	// Run Whisper to generate the transcript
	cmd := exec.Command("whisper", tmpFile.Name(), "--model", "large-v2", "--append_punctuations", "APPEND_PUNCTUATIONS", "--language", "en")
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Transcription failed: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	fmt.Println("output is", output)

	// Read the transcript from the output file
	transcript, err := ioutil.ReadFile("output.txt")
	if err != nil {
		http.Error(w, "Failed to read transcript", http.StatusInternalServerError)
		return
	}

	// Send the transcript as a response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(transcript)
}
