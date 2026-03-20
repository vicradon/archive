package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

// ── Shape definitions ──────────────────────────────────────────────────────────

type ShapeClass struct {
	ID    int
	Name  string
	Sides int
}

var classes = []ShapeClass{
	{0, "triangle", 3},
	{1, "quadrilateral", 4},
	{2, "pentagon", 5},
	{3, "hexagon", 6},
	{4, "heptagon", 7},
	{5, "octagon", 8},
	{6, "decagon", 10},
}

// ── Color palettes ─────────────────────────────────────────────────────────────

var fillColors = []color.NRGBA{
	{230, 57, 70, 255},
	{244, 162, 97, 255},
	{42, 157, 143, 255},
	{69, 123, 157, 255},
	{155, 93, 229, 255},
	{247, 37, 133, 255},
	{76, 201, 240, 255},
	{128, 185, 24, 255},
	{255, 107, 53, 255},
	{168, 218, 220, 255},
}

var bgColors = []color.NRGBA{
	{255, 255, 255, 255},
	{248, 249, 250, 255},
	{26, 26, 46, 255},
	{15, 52, 96, 255},
	{22, 33, 62, 255},
	{245, 245, 240, 255},
	{240, 237, 255, 255},
	{20, 30, 20, 255},
}

// ── Rendering ─────────────────────────────────────────────────────────────────

type Vec2 struct{ X, Y float64 }

func polygonVertices(cx, cy, r float64, sides int, rotation float64) []Vec2 {
	pts := make([]Vec2, sides)
	for i := 0; i < sides; i++ {
		angle := 2*math.Pi*float64(i)/float64(sides) + rotation
		pts[i] = Vec2{cx + r*math.Cos(angle), cy + r*math.Sin(angle)}
	}
	return pts
}

// Bresenham-style filled polygon using scanline
func fillPolygon(img *image.NRGBA, pts []Vec2, c color.NRGBA, alpha float64) {
	bounds := img.Bounds()
	minY, maxY := math.MaxFloat64, -math.MaxFloat64
	for _, p := range pts {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	n := len(pts)
	for y := int(math.Floor(minY)); y <= int(math.Ceil(maxY)); y++ {
		if y < bounds.Min.Y || y >= bounds.Max.Y {
			continue
		}
		var xs []float64
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			yi, yj := pts[i].Y, pts[j].Y
			xi, xj := pts[i].X, pts[j].X
			if (yi <= float64(y) && yj > float64(y)) || (yj <= float64(y) && yi > float64(y)) {
				t := (float64(y) - yi) / (yj - yi)
				xs = append(xs, xi+t*(xj-xi))
			}
		}
		// sort xs
		for i := 0; i < len(xs)-1; i++ {
			for j := i + 1; j < len(xs); j++ {
				if xs[j] < xs[i] {
					xs[i], xs[j] = xs[j], xs[i]
				}
			}
		}
		for i := 0; i+1 < len(xs); i += 2 {
			x0 := int(math.Ceil(xs[i]))
			x1 := int(math.Floor(xs[i+1]))
			for x := x0; x <= x1; x++ {
				if x < bounds.Min.X || x >= bounds.Max.X {
					continue
				}
				blendPixel(img, x, y, c, alpha)
			}
		}
	}
}

func blendPixel(img *image.NRGBA, x, y int, c color.NRGBA, alpha float64) {
	base := img.NRGBAAt(x, y)
	a := alpha
	r := uint8(float64(base.R)*(1-a) + float64(c.R)*a)
	g := uint8(float64(base.G)*(1-a) + float64(c.G)*a)
	b := uint8(float64(base.B)*(1-a) + float64(c.B)*a)
	img.SetNRGBA(x, y, color.NRGBA{r, g, b, 255})
}

// Draw anti-aliased line using Wu's algorithm
func drawLine(img *image.NRGBA, x0, y0, x1, y1 float64, c color.NRGBA, width float64) {
	dx := x1 - x0
	dy := y1 - y0
	length := math.Sqrt(dx*dx + dy*dy)
	if length < 0.001 {
		return
	}
	// Perpendicular
	nx, ny := -dy/length, dx/length
	hw := width / 2
	steps := int(math.Ceil(length)) + 1
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		cx := x0 + t*dx
		cy := y0 + t*dy
		// Draw across width
		for w := -hw; w <= hw; w += 0.5 {
			px := cx + nx*w
			py := cy + ny*w
			ix, iy := int(math.Round(px)), int(math.Round(py))
			bounds := img.Bounds()
			if ix >= bounds.Min.X && ix < bounds.Max.X && iy >= bounds.Min.Y && iy < bounds.Max.Y {
				blendPixel(img, ix, iy, c, 0.85)
			}
		}
	}
}

func strokePolygon(img *image.NRGBA, pts []Vec2, c color.NRGBA, width float64) {
	n := len(pts)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		drawLine(img, pts[i].X, pts[i].Y, pts[j].X, pts[j].Y, c, width)
	}
}

type BBox struct{ X, Y, W, H float64 }

func generateShapeImage(rng *rand.Rand, size, sides int) (*image.NRGBA, BBox) {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	// Background
	bg := bgColors[rng.Intn(len(bgColors))]
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	// Optional background noise
	if rng.Float64() > 0.6 {
		for i := 0; i < 25; i++ {
			nc := fillColors[rng.Intn(len(fillColors))]
			nx := rng.Float64() * float64(size)
			ny := rng.Float64() * float64(size)
			nr := rng.Float64() * float64(size) * 0.12
			circlePts := polygonVertices(nx, ny, nr, 12, 0)
			fillPolygon(img, circlePts, nc, 0.04)
		}
	}

	// Shape
	margin := math.Max(4, float64(size)*0.12)
	maxR := float64(size)/2 - margin
	minR := math.Max(4, float64(size)*0.18)
	r := minR + rng.Float64()*(maxR-minR)
	cx := margin + r + rng.Float64()*(float64(size)-2*margin-2*r)
	cy := margin + r + rng.Float64()*(float64(size)-2*margin-2*r)
	rotation := rng.Float64() * 2 * math.Pi

	fillIdx := rng.Intn(len(fillColors))
	fc := fillColors[fillIdx]
	fillAlpha := 0.55 + rng.Float64()*0.4

	pts := polygonVertices(cx, cy, r, sides, rotation)
	fillPolygon(img, pts, fc, fillAlpha)

	// Stroke — pick a different color
	strokeIdx := (fillIdx + 1 + rng.Intn(len(fillColors)-1)) % len(fillColors)
	sc := fillColors[strokeIdx]
	lw := math.Max(1.0, float64(size)*0.025)
	strokePolygon(img, pts, sc, lw)

	// Compute bounding box
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, p := range pts {
		if p.X < minX { minX = p.X }
		if p.Y < minY { minY = p.Y }
		if p.X > maxX { maxX = p.X }
		if p.Y > maxY { maxY = p.Y }
	}
	minX = math.Max(0, minX)
	minY = math.Max(0, minY)
	maxX = math.Min(float64(size), maxX)
	maxY = math.Min(float64(size), maxY)

	return img, BBox{minX, minY, maxX - minX, maxY - minY}
}

// ── Export formats ─────────────────────────────────────────────────────────────

type GeneratedImage struct {
	Data []byte
	BBox BBox
}

type ClassImages struct {
	Class  ShapeClass
	Images []GeneratedImage
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// COCO JSON structures
type CocoDataset struct {
	Info        CocoInfo        `json:"info"`
	Categories  []CocoCategory  `json:"categories"`
	Images      []CocoImage     `json:"images"`
	Annotations []CocoAnnotation `json:"annotations"`
}
type CocoInfo struct {
	Description string `json:"description"`
	Version     string `json:"version"`
}
type CocoCategory struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Supercategory string `json:"supercategory"`
}
type CocoImage struct {
	ID       int    `json:"id"`
	FileName string `json:"file_name"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}
type CocoAnnotation struct {
	ID         int       `json:"id"`
	ImageID    int       `json:"image_id"`
	CategoryID int       `json:"category_id"`
	BBox       [4]float64 `json:"bbox"`
	Area       float64   `json:"area"`
	IsCrowd    int       `json:"iscrowd"`
}

// VOC XML structures
type VocAnnotation struct {
	XMLName  xml.Name  `xml:"annotation"`
	Filename string    `xml:"filename"`
	Size     VocSize   `xml:"size"`
	Object   VocObject `xml:"object"`
}
type VocSize struct {
	Width  int `xml:"width"`
	Height int `xml:"height"`
	Depth  int `xml:"depth"`
}
type VocObject struct {
	Name    string   `xml:"name"`
	BndBox  VocBndBox `xml:"bndbox"`
}
type VocBndBox struct {
	Xmin int `xml:"xmin"`
	Ymin int `xml:"ymin"`
	Xmax int `xml:"xmax"`
	Ymax int `xml:"ymax"`
}

// ── ZIP writer helpers ─────────────────────────────────────────────────────────

func addToZip(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// ── Format builders ────────────────────────────────────────────────────────────

func writePlain(zw *zip.Writer, dataset []ClassImages) error {
	for _, cls := range dataset {
		for i, img := range cls.Images {
			name := fmt.Sprintf("%s/%s_%04d.png", cls.Class.Name, cls.Class.Name, i)
			if err := addToZip(zw, name, img.Data); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeCOCO(zw *zip.Writer, dataset []ClassImages, size int) error {
	coco := CocoDataset{
		Info: CocoInfo{Description: "2D Shape Dataset", Version: "1.0"},
	}
	for _, cls := range dataset {
		coco.Categories = append(coco.Categories, CocoCategory{
			ID: cls.Class.ID + 1, Name: cls.Class.Name, Supercategory: "shape",
		})
	}
	imgID, annID := 1, 1
	for _, cls := range dataset {
		for i, img := range cls.Images {
			fname := fmt.Sprintf("images/%s_%04d.png", cls.Class.Name, i)
			coco.Images = append(coco.Images, CocoImage{
				ID: imgID, FileName: fname, Width: size, Height: size,
			})
			bb := img.BBox
			coco.Annotations = append(coco.Annotations, CocoAnnotation{
				ID: annID, ImageID: imgID, CategoryID: cls.Class.ID + 1,
				BBox:    [4]float64{round2(bb.X), round2(bb.Y), round2(bb.W), round2(bb.H)},
				Area:    round2(bb.W * bb.H),
				IsCrowd: 0,
			})
			if err := addToZip(zw, fname, img.Data); err != nil {
				return err
			}
			imgID++
			annID++
		}
	}
	jsonBytes, err := json.MarshalIndent(coco, "", "  ")
	if err != nil {
		return err
	}
	return addToZip(zw, "annotations/instances.json", jsonBytes)
}

func writeYOLO(zw *zip.Writer, dataset []ClassImages, size int) error {
	var sb strings.Builder
	sb.WriteString("path: .\ntrain: images\nval: images\n")
	sb.WriteString(fmt.Sprintf("nc: %d\n", len(dataset)))
	names := make([]string, len(dataset))
	for i, cls := range dataset {
		names[i] = fmt.Sprintf("'%s'", cls.Class.Name)
	}
	sb.WriteString(fmt.Sprintf("names: [%s]\n", strings.Join(names, ", ")))
	if err := addToZip(zw, "data.yaml", []byte(sb.String())); err != nil {
		return err
	}
	for _, cls := range dataset {
		for i, img := range cls.Images {
			base := fmt.Sprintf("%s_%04d", cls.Class.Name, i)
			if err := addToZip(zw, "images/"+base+".png", img.Data); err != nil {
				return err
			}
			bb := img.BBox
			f := float64(size)
			cx := (bb.X + bb.W/2) / f
			cy := (bb.Y + bb.H/2) / f
			nw := bb.W / f
			nh := bb.H / f
			label := fmt.Sprintf("%d %.6f %.6f %.6f %.6f\n", cls.Class.ID, cx, cy, nw, nh)
			if err := addToZip(zw, "labels/"+base+".txt", []byte(label)); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeVOC(zw *zip.Writer, dataset []ClassImages, size int) error {
	for _, cls := range dataset {
		for i, img := range cls.Images {
			base := fmt.Sprintf("%s_%04d", cls.Class.Name, i)
			if err := addToZip(zw, "JPEGImages/"+base+".png", img.Data); err != nil {
				return err
			}
			bb := img.BBox
			ann := VocAnnotation{
				Filename: base + ".png",
				Size:     VocSize{Width: size, Height: size, Depth: 3},
				Object: VocObject{
					Name: cls.Class.Name,
					BndBox: VocBndBox{
						Xmin: int(math.Round(bb.X)),
						Ymin: int(math.Round(bb.Y)),
						Xmax: int(math.Round(bb.X + bb.W)),
						Ymax: int(math.Round(bb.Y + bb.H)),
					},
				},
			}
			xmlBytes, err := xml.MarshalIndent(ann, "", "  ")
			if err != nil {
				return err
			}
			xmlBytes = append([]byte(xml.Header), xmlBytes...)
			if err := addToZip(zw, "Annotations/"+base+".xml", xmlBytes); err != nil {
				return err
			}
		}
	}
	return nil
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

// ── Progress bar ───────────────────────────────────────────────────────────────

func printProgress(label string, done, total int) {
	pct := float64(done) / float64(total)
	barW := 40
	filled := int(pct * float64(barW))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barW-filled)
	fmt.Printf("\r  %s [%s] %3d%%  %d/%d", label, bar, int(pct*100), done, total)
}

// ── Main ───────────────────────────────────────────────────────────────────────

func main() {
	// Flags
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
		for _, c := range classes {
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
	activeClasses := classes
	if *classesF != "all" {
		names := strings.Split(*classesF, ",")
		nameSet := make(map[string]bool)
		for _, n := range names {
			nameSet[strings.TrimSpace(strings.ToLower(n))] = true
		}
		activeClasses = nil
		for _, c := range classes {
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
			for i, c := range activeClasses { ns[i] = c.Name }
			return strings.Join(ns, ", ")
		}(),
		*perClass, total,
		*size, *size,
		strings.ToUpper(*format),
		*workers,
		outPath)

	// ── Generate ────────────────────────────────────────────────────────────────
	fmt.Println("  Generating images…")
	startGen := time.Now()

	dataset := make([]ClassImages, len(activeClasses))
	var mu sync.Mutex
	var wg sync.WaitGroup

	done := 0

	for ci, cls := range activeClasses {
		dataset[ci] = ClassImages{Class: cls, Images: make([]GeneratedImage, *perClass)}

		// Split work into chunks per worker
		chunk := (*perClass + *workers - 1) / *workers
		for w := 0; w < *workers; w++ {
			start := w * chunk
			end := start + chunk
			if start >= *perClass { break }
			if end > *perClass { end = *perClass }

			wg.Add(1)
			go func(ci, start, end int, cls ShapeClass, workerSeed int64) {
				defer wg.Done()
				rng := rand.New(rand.NewSource(workerSeed))
				for i := start; i < end; i++ {
					img, bbox := generateShapeImage(rng, *size, cls.Sides)
					pngData, _ := encodePNG(img)
					dataset[ci].Images[i] = GeneratedImage{Data: pngData, BBox: bbox}

					mu.Lock()
					done++
					if done%50 == 0 || done == total {
						printProgress(fmt.Sprintf("%-14s", cls.Name), done, total)
					}
					mu.Unlock()
				}
			}(ci, start, end, cls, baseSeed+int64(ci*1000+w))
		}
	}
	wg.Wait()
	printProgress("done", total, total)
	fmt.Printf("\n  ✓ Generated %d images in %s\n\n", total, time.Since(startGen).Round(time.Millisecond))

	// ── Pack ZIP ────────────────────────────────────────────────────────────────
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
		err = writePlain(zw, dataset)
	case "coco":
		err = writeCOCO(zw, dataset, *size)
	case "yolo":
		err = writeYOLO(zw, dataset, *size)
	case "voc":
		err = writeVOC(zw, dataset, *size)
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
