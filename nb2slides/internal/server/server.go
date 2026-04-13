package server

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"nb2slides/internal/slides"
)

// ServePreviews starts a preview server for the slides
func ServePreviews(slideList []slides.Slide, port string) {
	html := fullHTML(slideList, "Notebook Preview")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	addr := "http://localhost:" + port
	fmt.Printf("Preview server running at %s\n", addr)
	fmt.Println("Open that URL in your browser, then Ctrl+P → Save as PDF.")
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

func fullHTML(slideList []slides.Slide, title string) string {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html lang="en"><head>`)
	b.WriteString(`<meta charset="UTF-8">`)
	b.WriteString(fmt.Sprintf(`<title>%s</title>`, title))
	b.WriteString(`<style>* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: Georgia, 'Times New Roman', serif; font-size: 13px; background: #fff; color: #111; }
@media screen { body { background: #f2f2f2; padding: 28px 20px; }
.slide { width: 210mm; min-height: 148mm; margin: 0 auto 20px; background: #ffffff; border: 1px solid #d8d8d8; border-radius: 3px; box-shadow: 0 1px 4px rgba(0,0,0,0.08); padding: 14mm 18mm 13mm; page-break-after: always; position: relative; } }
@media print { body { background: #fff; margin: 0; padding: 0; }
.slide { width: 100%; padding: 14mm 18mm 13mm; page-break-after: always; break-after: page; position: relative; border: none; box-shadow: none; }
.export-bar { display: none !important; } }
.slide-label { position: absolute; top: 7mm; right: 10mm; font-size: 9.5px; color: #bbb; font-family: 'Helvetica Neue', Arial, sans-serif; letter-spacing: 0.4px; }
.cell-md { font-family: Georgia, 'Times New Roman', serif; font-size: 13.5px; line-height: 1.65; color: #1a1a1a; }
.cell-md h1 { font-size: 20px; font-weight: 700; border-bottom: 1.5px solid #555; padding-bottom: 5px; margin-bottom: 12px; font-family: 'Helvetica Neue', Arial, sans-serif; }
.cell-md h2 { font-size: 16px; font-weight: 600; border-bottom: 1px solid #bbb; padding-bottom: 4px; margin-bottom: 9px; font-family: 'Helvetica Neue', Arial, sans-serif; }
.cell-md h3 { font-size: 14px; font-weight: 600; margin-bottom: 6px; font-family: 'Helvetica Neue', Arial, sans-serif; }
.cell-md p { margin-bottom: 7px; }
.cell-md ul, .cell-md ol { margin-left: 22px; margin-bottom: 7px; }
.cell-md li { margin-bottom: 3px; }
.cell-md code { background: #f5f5f5; border: 1px solid #e0e0e0; padding: 1px 4px; border-radius: 3px; font-size: 11.5px; font-family: 'Courier New', Courier, monospace; }
.cell-md pre { background: #f5f5f5; border: 1px solid #e0e0e0; padding: 8px 10px; border-radius: 3px; margin-bottom: 8px; overflow-x: auto; }
.cell-md pre code { border: none; background: none; padding: 0; }
.cell-md blockquote { border-left: 3px solid #ccc; padding-left: 10px; color: #555; margin-bottom: 7px; }
.cell-md .MJX-TEX { font-size: 13px; }
.mjx-chtml { margin: 6px 0 !important; }
.math-display { margin: 8px 0; overflow-x: auto; }
.cell-code { background: #fafafa; border-left: 3px solid #5b8dd9; border-radius: 0 3px 3px 0; }
.cell-code pre { padding: 9px 12px; overflow-x: auto; white-space: pre-wrap; word-break: break-all; }
.cell-code code { font-family: 'Courier New', Courier, monospace; font-size: 12px; line-height: 1.55; color: #1a1a1a; }
.cell-output { margin-top: 9px; border-top: 1px dashed #ddd; padding-top: 7px; }
.output-label { font-size: 9.5px; color: #aaa; font-family: 'Helvetica Neue', Arial, sans-serif; text-transform: uppercase; letter-spacing: 0.6px; margin-bottom: 4px; }
.out-stdout { background: #f7f7f7; border-radius: 3px; padding: 5px 8px; font-family: 'Courier New', Courier, monospace; font-size: 11.5px; line-height: 1.5; white-space: pre-wrap; word-break: break-all; color: #222; }
.out-stderr { background: #fff8f8; border-left: 3px solid #d44; padding: 5px 8px; font-family: 'Courier New', Courier, monospace; font-size: 11.5px; line-height: 1.5; white-space: pre-wrap; word-break: break-all; color: #800; }
.out-image img { max-width: 100%; max-height: 220px; display: block; margin: 4px 0; }
.out-html { font-size: 11.5px; overflow-x: auto; }
.out-html table { border-collapse: collapse; font-size: 11px; }
.out-html td, .out-html th { border: 1px solid #ddd; padding: 3px 7px; }
.out-html th { background: #f3f3f3; font-weight: 600; }
.cell-raw pre { background: #fffdf0; border: 1px solid #e8d44d; padding: 8px 10px; border-radius: 3px; font-family: 'Courier New', Courier, monospace; font-size: 11.5px; white-space: pre-wrap; }
.export-bar { position: fixed; top: 14px; right: 18px; z-index: 100; }
.export-bar button { padding: 6px 16px; font-size: 13px; border: 1px solid #888; border-radius: 4px; background: #fff; cursor: pointer; box-shadow: 0 1px 3px rgba(0,0,0,0.12); font-family: 'Helvetica Neue', Arial, sans-serif; }
.export-bar button:hover { background: #f5f5f5; }</style>`)
	b.WriteString(`<script>window.MathJax = { tex: { inlineMath: [['$', '$']], displayMath: [['$$', '$$'], ['\\\\[', '\\\\]']], processEscapes: true, packages: {'[+]': ['ams']} }, options: { skipHtmlTags: ['script','noscript','style','textarea','pre','code'] }, startup: { typeset: true } };</script>`)
	b.WriteString(`<script defer src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-chtml.js"></script>`)
	b.WriteString(`</head><body>`)
	b.WriteString(`<div class="export-bar">`)
	b.WriteString(`<button onclick="window.print()">🖨 Print / Export PDF</button>`)
	b.WriteString(`</div>`)
	for _, s := range slideList {
		b.WriteString(s.HTML)
	}
	b.WriteString(`</body></html>`)
	return b.String()
}
