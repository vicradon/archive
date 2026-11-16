package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"shapegen/export"
	"shapegen/progress"
	"shapegen/render"
	"shapegen/shapes"
)

func main() {
	size     := flag.Int("size", 40, "Image size in pixels (square)")
	perClass := flag.Int("n", 100, "Images per class")
	format   := flag.String("format", "coco", "Export format: plain, coco, yolo, voc")
	output   := flag.String("o", "", "Output ZIP file (default: shape_dataset_<format>.zip)")
	classesF := flag.String("classes", "all", "Comma-separated class names or 'all'")
	workers  := flag.Int("workers", 4, "Parallel workers for generation")
	seed     := flag.Int64("seed", 0, "Random seed (0 = time-based)")
	list     := flag.Bool("list", false, "List available classes and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
╔══════════════════════════════════════════════════════════╗
║          Shape Dataset Generator  —  CLI                 ║
╚══════════════════════════════════════════════════════════╝

Generates synthetic 2D polygon images for CV training.
Classes: triangle, quadrilateral, pentagon, hexagon,
         heptagon, octagon, decagon

Usage:
  shapegen [flags]

Flags:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Examples:
  shapegen
  shapegen -size 64 -n 200 -format yolo -o my_dataset.zip
  shapegen -classes triangle,hexagon,decagon -n 50 -format plain
  shapegen -size 128 -n 500 -workers 8 -format coco
`)
	}
	flag.Parse()

	if *list {
		fmt.Println("\nAvailable shape classes:")
		for _, c := range shapes.Classes {
			fmt.Printf("  %-16s  (%d sides)\n", c.Name, c.Sides)
		}
		return
	}

	// Validate format
	validFormats := map[string]bool{"plain": true, "coco": true, "yolo": true, "voc": true}
	if !validFormats[*format] {
		fmt.Fprintf(os.Stderr, "error: unknown format %q — choose from plain, coco, yolo, voc\n", *format)
		os.Exit(1)
	}

	// Parse classes
	activeClasses := shapes.Classes
	if *classesF != "all" {
		names := strings.Split(*classesF, ",")
		nameSet := make(map[string]bool)
		for _, n := range names {
			nameSet[strings.TrimSpace(strings.ToLower(n))] = true
		}
		activeClasses = nil
		for _, c := range shapes.Classes {
			if nameSet[c.Name] {
				activeClasses = append(activeClasses, c)
			}
		}
		if len(activeClasses) == 0 {
			fmt.Fprintf(os.Stderr, "error: no valid classes matched %q\n", *classesF)
			os.Exit(1)
		}
	}

	// Output path
	outPath := *output
	if outPath == "" {
		outPath = fmt.Sprintf("shape_dataset_%s.zip", *format)
	}

	// Seed
	baseSeed := *seed
	if baseSeed == 0 {
		baseSeed = time.Now().UnixNano()
	}

	total := len(activeClasses) * (*perClass)

	// Banner
	fmt.Printf(`
  ╔═══════════════════════════════════════╗
  ║    Shape Dataset Generator  v1.0      ║
  ╚═══════════════════════════════════════╝

  Classes  : %d  (%s)
  Images   : %d per class  (%d total)
  Size     : %dx%d px
  Format   : %s
  Workers  : %d
  Output   : %s

`, len(activeClasses),
		func() string {
			ns := make([]string, len(activeClasses))
			for i, c := range activeClasses {
				ns[i] = c.Name
			}
			return strings.Join(ns, ", ")
		}(),
		*perClass, total,
		*size, *size,
		strings.ToUpper(*format),
		*workers,
		outPath)

	// ── Generate ─────────────────────────────────────────────────────────────────
	fmt.Println("  Generating images…")
	startGen := time.Now()

	dataset := make([]render.ClassImages, len(activeClasses))
	var mu sync.Mutex
	var wg sync.WaitGroup
	done := 0

	for ci, cls := range activeClasses {
		dataset[ci] = render.ClassImages{Class: cls, Images: make([]render.GeneratedImage, *perClass)}

		chunk := (*perClass + *workers - 1) / *workers
		for w := 0; w < *workers; w++ {
			start := w * chunk
			end := start + chunk
			if start >= *perClass {
				break
			}
			if end > *perClass {
				end = *perClass
			}

			wg.Add(1)
			go func(ci, start, end int, cls shapes.ShapeClass, workerSeed int64) {
				defer wg.Done()
				rng := rand.New(rand.NewSource(workerSeed))
				for i := start; i < end; i++ {
					img, bbox := render.GenerateShapeImage(rng, *size, cls.Sides)
					pngData, _ := export.EncodePNG(img)
					dataset[ci].Images[i] = render.GeneratedImage{Data: pngData, BBox: bbox}

					mu.Lock()
					done++
					if done%50 == 0 || done == total {
						progress.Print(fmt.Sprintf("%-14s", cls.Name), done, total)
					}
					mu.Unlock()
				}
			}(ci, start, end, cls, baseSeed+int64(ci*1000+w))
		}
	}
	wg.Wait()
	progress.Print("done", total, total)
	fmt.Printf("\n  ✓ Generated %d images in %s\n\n", total, time.Since(startGen).Round(time.Millisecond))

	// ── Pack ZIP ──────────────────────────────────────────────────────────────────
	fmt.Printf("  Packing %s ZIP…\n", strings.ToUpper(*format))
	startZip := time.Now()

	f, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	switch *format {
	case "plain":
		err = export.WritePlain(zw, dataset)
	case "coco":
		err = export.WriteCOCO(zw, dataset, *size)
	case "yolo":
		err = export.WriteYOLO(zw, dataset, *size)
	case "voc":
		err = export.WriteVOC(zw, dataset, *size)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing zip: %v\n", err)
		os.Exit(1)
	}
	zw.Close()

	fi, _ := os.Stat(outPath)
	zipSize := ""
	if fi != nil {
		b := fi.Size()
		switch {
		case b >= 1<<20:
			zipSize = fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
		case b >= 1<<10:
			zipSize = fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
		default:
			zipSize = fmt.Sprintf("%d B", b)
		}
	}

	fmt.Printf("  ✓ ZIP written in %s  (%s)\n\n", time.Since(startZip).Round(time.Millisecond), zipSize)
	fmt.Printf("  ┌─────────────────────────────────────┐\n")
	fmt.Printf("  │  Output → %-26s│\n", outPath+" ")
	fmt.Printf("  └─────────────────────────────────────┘\n\n")
}
