package main

import (
	"flag"
	"fmt"
	"os"

	"nb2slides/internal/export"
	"nb2slides/internal/notebook"
	"nb2slides/internal/render"
	"nb2slides/internal/server"
	"nb2slides/internal/slides"
)

func main() {
	var (
		inputFile   = flag.String("i", "", "Input .ipynb file (required)")
		outputDir   = flag.String("o", "slides_output", "Output directory for PDF slides")
		maxLines    = flag.Int("max-lines", 50, "Max lines per slide before splitting (default 50)")
		showOutputs = flag.Bool("outputs", true, "Include cell outputs (default true)")
		serve       = flag.Bool("serve", false, "Open browser preview instead of exporting (useful when chromium is unavailable)")
		port        = flag.String("port", "8080", "Port for preview server (used with -serve)")
		themeFlag   = flag.String("theme", "dark", "Colour theme: white | dark | sepia")
	)
	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintln(os.Stderr, "Error: -i <notebook.ipynb> is required")
		flag.Usage()
		os.Exit(1)
	}

	theme := render.Theme(*themeFlag)
	switch theme {
	case render.ThemeWhite, render.ThemeDark, render.ThemeSepia:
		// valid
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown theme %q — must be white, dark, or sepia\n", *themeFlag)
		os.Exit(1)
	}

	nb, err := notebook.ParseNotebook(*inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading notebook: %v\n", err)
		os.Exit(1)
	}

	slideList := slides.BuildSlides(nb, *maxLines, *showOutputs)
	fmt.Printf("Notebook parsed: %d cells → %d slides\n", len(nb.Cells), len(slideList))

	if *serve {
		server.ServePreviews(slideList, *port, theme)
		return
	}

	if err := export.ExportPDFs(slideList, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Done. PDFs written to: %s/\n", *outputDir)
}
