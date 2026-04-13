package export

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"nb2slides/internal/render"
	"nb2slides/internal/slides"
)

// ChromiumBin tries to find chromium / google-chrome on PATH
func ChromiumBin() string {
	for _, name := range []string{"chromium-browser", "chromium", "google-chrome", "google-chrome-stable", "chrome"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

// ExportPDFs exports slides to PDF files
func ExportPDFs(slides []slides.Slide, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	chrome := ChromiumBin()
	if chrome == "" {
		return exportHTMLFallback(slides, outDir)
	}

	fmt.Printf("Using browser: %s\n", chrome)

	// Write combined HTML for all-in-one option too
	combinedPath := filepath.Join(outDir, "all_slides.html")
	combined := fullHTML(slides, "Notebook Slides")
	if err := os.WriteFile(combinedPath, []byte(combined), 0644); err != nil {
		return err
	}
	fmt.Printf("Combined HTML: %s\n", combinedPath)

	// Export each slide as its own PDF
	tmpDir, err := os.MkdirTemp("", "nb2slides-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	for _, s := range slides {
		htmlPath := filepath.Join(tmpDir, fmt.Sprintf("slide_%s.html", safeName(s.ID)))
		pdfPath := filepath.Join(outDir, fmt.Sprintf("slide_%s.pdf", safeName(s.ID)))

		html := singleSlideHTML(s, "Notebook")
		if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
			return err
		}

		args := []string{
			"--headless",
			"--disable-gpu",
			"--no-sandbox",
			"--print-to-pdf=" + pdfPath,
			"--print-to-pdf-no-header",
			"--no-margins",
			"file://" + htmlPath,
		}
		cmd := exec.Command(chrome, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: chrome failed for slide %s: %v\n", s.ID, err)
		} else {
			fmt.Printf("  ✓ slide_%s.pdf\n", s.ID)
		}
	}

	return nil
}

// exportHTMLFallback writes per-slide HTML files when chromium is not available
func exportHTMLFallback(slides []slides.Slide, outDir string) error {
	fmt.Println("Chromium not found. Exporting as HTML files (open in browser → Print → Save as PDF).")

	// Write one HTML per slide
	for _, s := range slides {
		name := fmt.Sprintf("slide_%s.html", safeName(s.ID))
		path := filepath.Join(outDir, name)
		html := singleSlideHTML(s, "Notebook")
		if err := os.WriteFile(path, []byte(html), 0644); err != nil {
			return err
		}
	}

	// Write combined HTML
	combined := fullHTML(slides, "Notebook Slides")
	combinedPath := filepath.Join(outDir, "all_slides.html")
	if err := os.WriteFile(combinedPath, []byte(combined), 0644); err != nil {
		return err
	}

	// Write an index.html that lists all slides
	var idx strings.Builder
	idx.WriteString(`<!DOCTYPE html><html><head><meta charset="UTF-8">`)
	idx.WriteString(`<title>Slide Index</title>`)
	idx.WriteString(`<style>body{font-family:sans-serif;padding:20px}a{display:block;padding:4px 0;font-size:14px}`)
	idx.WriteString(`.tip{background:#fffbe6;border:1px solid #e8d44d;padding:12px;border-radius:4px;margin-bottom:16px}</style>`)
	idx.WriteString(`</head><body>`)
	idx.WriteString(`<div class="tip">`)
	idx.WriteString(`<strong>To export as PDF:</strong> Open <a href="all_slides.html">all_slides.html</a>, `)
	idx.WriteString(`then use <kbd>Ctrl+P</kbd> (or File → Print) → Save as PDF → choose "No margins", "Background graphics on".</div>`)
	idx.WriteString(`<h2>Slides</h2>`)
	for _, s := range slides {
		name := fmt.Sprintf("slide_%s.html", safeName(s.ID))
		idx.WriteString(fmt.Sprintf(`<a href="%s">Slide %s (%s)</a>`, name, s.ID, s.CellType))
	}
	idx.WriteString(`</body></html>`)
	return os.WriteFile(filepath.Join(outDir, "index.html"), []byte(idx.String()), 0644)
}

func fullHTML(slideList []slides.Slide, title string) string {
	var b strings.Builder
	b.WriteString(render.HTMLHead(title, ""))
	b.WriteString(`<body>`)

	b.WriteString(`<div class="export-bar">`)
	b.WriteString(`<button onclick="window.print()">🖨 Print / Export PDF</button>`)
	b.WriteString(`</div>`)

	for _, s := range slideList {
		b.WriteString(s.HTML)
	}

	b.WriteString(`</body></html>`)
	return b.String()
}

func singleSlideHTML(s slides.Slide, title string) string {
	extra := `
@page { size: A4 landscape; margin: 0; }
@media print { .slide { min-height: 100vh; } }
`
	var b strings.Builder
	b.WriteString(render.HTMLHead(fmt.Sprintf("%s – %s", title, s.ID), extra))
	b.WriteString(`<body>`)
	b.WriteString(s.HTML)
	b.WriteString(`</body></html>`)
	return b.String()
}

func safeName(id string) string {
	return strings.ReplaceAll(id, "-", "_")
}
