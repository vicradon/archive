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
	"strconv"
	"sync"
)

const uploadDir = "./uploads"

type ByteRange struct {
	Start int64
	End   int64
}

type UploadSession struct {
	File           *os.File
	TotalSize      int64
	ReceivedRanges []ByteRange
	Paused         bool
	Canceled       bool
	mu             sync.Mutex
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
	http.HandleFunc("/status", statusHandler)

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
	offsetStr := r.FormValue("offset")
	totalSizeStr := r.FormValue("totalSize")

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid offset", http.StatusBadRequest)
		return
	}

	totalSize, err := strconv.ParseInt(totalSizeStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid total size", http.StatusBadRequest)
		return
	}

	chunk, header, err := r.FormFile("chunk")
	if err != nil {
		http.Error(w, "Error reading chunk", http.StatusBadRequest)
		return
	}
	defer chunk.Close()

	chunkSize := header.Size

	sessionID := r.RemoteAddr + filename
	sessionsMu.Lock()
	session, exists := sessions[sessionID]
	if !exists {
		filepath := filepath.Join(uploadDir, filename)

		// Open file with O_RDWR to allow seeking and writing
		file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			sessionsMu.Unlock()
			http.Error(w, "Error creating file", http.StatusInternalServerError)
			return
		}

		// Pre-allocate file to total size to prevent fragmentation
		if err := file.Truncate(totalSize); err != nil {
			file.Close()
			sessionsMu.Unlock()
			http.Error(w, "Error allocating file space", http.StatusInternalServerError)
			return
		}

		session = &UploadSession{
			File:           file,
			TotalSize:      totalSize,
			ReceivedRanges: make([]ByteRange, 0),
		}
		sessions[sessionID] = session
	}
	sessionsMu.Unlock()

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.Canceled {
		http.Error(w, "Upload canceled", http.StatusBadRequest)
		return
	}

	// Check if this range was already received
	if isRangeReceived(session.ReceivedRanges, offset, offset+chunkSize) {
		log.Printf("Chunk at offset %d already received, skipping", offset)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "ok",
			"message":  "chunk already received",
			"progress": calculateProgress(session),
		})
		return
	}

	// Seek to the correct position
	if _, err := session.File.Seek(offset, 0); err != nil {
		http.Error(w, "Error seeking in file", http.StatusInternalServerError)
		return
	}

	// Verify chunk size matches what we expect
	actualChunkSize := header.Size

	// Read chunk into buffer first to ensure complete read
	buffer := make([]byte, actualChunkSize)
	n, err := io.ReadFull(chunk, buffer)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		log.Printf("Error reading chunk data: %v", err)
		http.Error(w, "Error reading chunk data", http.StatusInternalServerError)
		return
	}

	if int64(n) != actualChunkSize {
		log.Printf("WARNING: Expected to read %d bytes but got %d bytes", actualChunkSize, n)
	}

	// Write the exact amount read
	written, err := session.File.WriteAt(buffer[:n], offset)
	if err != nil {
		log.Printf("Error writing chunk at offset %d: %v", offset, err)
		http.Error(w, "Error writing chunk", http.StatusInternalServerError)
		return
	}

	if int64(written) != int64(n) {
		log.Printf("ERROR: Incomplete chunk write at offset %d: expected %d, wrote %d", offset, n, written)
		http.Error(w, fmt.Sprintf("Incomplete chunk write: expected %d, wrote %d", n, written), http.StatusInternalServerError)
		return
	}

	// Sync to ensure data is written to disk immediately
	if err := session.File.Sync(); err != nil {
		log.Printf("Error syncing file: %v", err)
		http.Error(w, "Error syncing file", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully wrote chunk at offset %d, size %d bytes (chunk %s/%s)", offset, written, chunkIndex, totalChunks)

	// Track received range
	newRange := ByteRange{Start: offset, End: offset + chunkSize}
	session.ReceivedRanges = mergeRanges(session.ReceivedRanges, newRange)

	progress := calculateProgress(session)

	log.Printf("Chunk %s/%s written at offset %d, size %d bytes, progress: %.2f%%",
		chunkIndex, totalChunks, offset, chunkSize, progress*100)

	// Check if upload is complete
	if isUploadComplete(session) {
		log.Printf("Upload complete for %s", filename)
		session.File.Close()
		sessionsMu.Lock()
		delete(sessions, sessionID)
		sessionsMu.Unlock()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "complete",
			"progress": 1.0,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "ok",
		"progress": progress,
	})
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

func statusHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	sessionID := r.RemoteAddr + filename
	sessionsMu.RLock()
	session, exists := sessions[sessionID]
	sessionsMu.RUnlock()

	if !exists {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"exists":   false,
			"progress": 0,
		})
		return
	}

	session.mu.Lock()
	progress := calculateProgress(session)
	ranges := make([]map[string]int64, len(session.ReceivedRanges))
	for i, r := range session.ReceivedRanges {
		ranges[i] = map[string]int64{"start": r.Start, "end": r.End}
	}
	session.mu.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"exists":   true,
		"progress": progress,
		"ranges":   ranges,
	})
}

// Helper functions

func isRangeReceived(ranges []ByteRange, start, end int64) bool {
	for _, r := range ranges {
		if start >= r.Start && end <= r.End {
			return true
		}
	}
	return false
}

func mergeRanges(ranges []ByteRange, newRange ByteRange) []ByteRange {
	ranges = append(ranges, newRange)

	// Sort ranges by start position
	for i := 0; i < len(ranges); i++ {
		for j := i + 1; j < len(ranges); j++ {
			if ranges[j].Start < ranges[i].Start {
				ranges[i], ranges[j] = ranges[j], ranges[i]
			}
		}
	}

	// Merge overlapping ranges
	merged := make([]ByteRange, 0)
	for _, r := range ranges {
		if len(merged) == 0 {
			merged = append(merged, r)
			continue
		}

		last := &merged[len(merged)-1]
		if r.Start <= last.End {
			// Overlapping or adjacent, merge
			if r.End > last.End {
				last.End = r.End
			}
		} else {
			merged = append(merged, r)
		}
	}

	return merged
}

func calculateProgress(session *UploadSession) float64 {
	if session.TotalSize == 0 {
		return 0
	}

	var received int64
	for _, r := range session.ReceivedRanges {
		received += r.End - r.Start
	}

	return float64(received) / float64(session.TotalSize)
}

func isUploadComplete(session *UploadSession) bool {
	if len(session.ReceivedRanges) != 1 {
		return false
	}

	r := session.ReceivedRanges[0]
	return r.Start == 0 && r.End == session.TotalSize
}
