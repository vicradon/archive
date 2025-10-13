package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const uploadDir = "./uploads"

type UploadSession struct {
	File     *os.File
	Paused   bool
	Canceled bool
	mu       sync.Mutex
}

var sessions = make(map[string]*UploadSession)
var sessionsMu sync.RWMutex

var tmpl *template.Template

type PageData struct {
	Message     string
	MessageType string
}

func main() {
	var err error
	tmpl, err = template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatalf("Error loading template: %v", err)
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload-chunk", uploadChunkHandler)
	http.HandleFunc("/cancel", cancelHandler)

	fmt.Println("Server starting on http://0.0.0.0:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, PageData{})
}

func uploadChunkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	filename := r.FormValue("filename")
	chunkIndex := r.FormValue("chunkIndex")
	totalChunks := r.FormValue("totalChunks")

	chunk, _, err := r.FormFile("chunk")
	if err != nil {
		http.Error(w, "Error reading chunk", http.StatusBadRequest)
		return
	}
	defer chunk.Close()

	sessionID := r.RemoteAddr + filename
	sessionsMu.Lock()
	session, exists := sessions[sessionID]
	if !exists {
		filepath := filepath.Join(uploadDir, filename)
		file, err := os.Create(filepath)
		if err != nil {
			sessionsMu.Unlock()
			http.Error(w, "Error creating file", http.StatusInternalServerError)
			return
		}
		session = &UploadSession{File: file}
		sessions[sessionID] = session
	}
	sessionsMu.Unlock()

	session.mu.Lock()
	if session.Canceled {
		session.mu.Unlock()
		http.Error(w, "Upload canceled", http.StatusBadRequest)
		return
	}

	if _, err := io.Copy(session.File, chunk); err != nil {
		session.mu.Unlock()
		http.Error(w, "Error writing chunk", http.StatusInternalServerError)
		return
	}
	session.mu.Unlock()

	if chunkIndex == totalChunks || chunkIndex+"1" == totalChunks {
		sessionsMu.Lock()
		session.File.Close()
		delete(sessions, sessionID)
		sessionsMu.Unlock()
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func cancelHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.RemoteAddr
	sessionsMu.Lock()
	for id, session := range sessions {
		if len(id) > len(sessionID) && id[:len(sessionID)] == sessionID {
			session.mu.Lock()
			session.Canceled = true
			session.File.Close()
			os.Remove(session.File.Name())
			session.mu.Unlock()
			delete(sessions, id)
		}
	}
	sessionsMu.Unlock()

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
