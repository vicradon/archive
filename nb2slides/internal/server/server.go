package server

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"nb2slides/internal/render"
	"nb2slides/internal/slides"
)

// ServePreviews starts a live preview server for the slides.
// theme controls the initial colour scheme; pass "" to use the default (white).
func ServePreviews(slideList []slides.Slide, port string, theme render.Theme) {
	html := fullHTML(slideList, "Notebook Preview", theme)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	addr := "http://localhost:" + port
	fmt.Printf("Preview server running at %s\n", addr)
	fmt.Println("Open that URL in your browser, then use the Print button to export PDF.")
	openBrowser(addr)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		// Linux: try xdg-open, then common browsers
		for _, b := range []string{"xdg-open", "chromium-browser", "chromium", "google-chrome", "firefox"} {
			if _, err := exec.LookPath(b); err == nil {
				cmd = exec.Command(b, url)
				break
			}
		}
	}
	if cmd != nil {
		_ = cmd.Start()
	}
}

func fullHTML(slideList []slides.Slide, title string, theme render.Theme) string {
	var b strings.Builder

	// HTMLHead opens <html>, <head>, and <body class="theme-*">
	b.WriteString(render.HTMLHead(title, theme, ""))

	// Fixed export + theme-toggle bar
	b.WriteString(render.ExportBar(theme))

	for _, s := range slideList {
		b.WriteString(s.HTML)
	}

	b.WriteString(`</body></html>`)
	return b.String()
}
