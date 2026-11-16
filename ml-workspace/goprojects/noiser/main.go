package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	T         = 1000
	betaStart = 1e-4
	betaEnd   = 0.02
)

// precompute alpha_bar for all timesteps
func buildSchedule() []float64 {
	alphaBar := make([]float64, T)
	product := 1.0
	for t := 0; t < T; t++ {
		beta := betaStart + float64(t)*(betaEnd-betaStart)/float64(T-1)
		product *= (1.0 - beta)
		alphaBar[t] = product
	}
	return alphaBar
}

func loadImageFromDisk(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return decodeImage(f, path)
}

func loadImageFromURL(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return decodeImage(resp.Body, url)
}

func decodeImage(r io.Reader, name string) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Decode(r)
	case ".png":
		return png.Decode(r)
	default:
		img, _, err := image.Decode(r)
		return img, err
	}
}

// apply forward diffusion: x_t = sqrt(alpha_bar_t)*x0 + sqrt(1-alpha_bar_t)*eps
func noiseStep(img image.Image, t int, alphaBar []float64) image.Image {
	a := alphaBar[t]
	sqrtA := math.Sqrt(a)
	sqrtOneMinusA := math.Sqrt(1 - a)

	bounds := img.Bounds()
	// NRGBA = non-premultiplied alpha, so pixel values stay in true [0,255] range
	// and we don't have to undo premultiplication before doing math
	out := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// RGBA() returns premultiplied 16-bit values [0,65535]
			// for images with no alpha (JPEG), alpha=65535 so divide is a no-op
			// for images with alpha we undo premultiplication manually
			pr, pg, pb, pa := img.At(x, y).RGBA()

			var rf, gf, bf float64
			if pa == 0 {
				rf, gf, bf = 0, 0, 0
			} else {
				// undo premultiplication: true channel = premultiplied / alpha * 65535
				rf = float64(pr)/float64(pa)*2 - 1
				gf = float64(pg)/float64(pa)*2 - 1
				bf = float64(pb)/float64(pa)*2 - 1
			}

			// sample independent gaussian noise per channel ~ N(0,1)
			nr := rand.NormFloat64()
			ng := rand.NormFloat64()
			nb := rand.NormFloat64()

			// forward diffusion: x_t = sqrt(a_bar)*x0 + sqrt(1-a_bar)*eps
			rf = sqrtA*rf + sqrtOneMinusA*nr
			gf = sqrtA*gf + sqrtOneMinusA*ng
			bf = sqrtA*bf + sqrtOneMinusA*nb

			out.SetNRGBA(x, y, color.NRGBA{
				R: floatToUint8(rf),
				G: floatToUint8(gf),
				B: floatToUint8(bf),
				A: uint8(pa >> 8), // preserve original alpha
			})
		}
	}
	return out
}

func floatToUint8(v float64) uint8 {
	v = (v + 1) / 2 * 255
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func saveImage(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 95})
	default:
		return png.Encode(f, img)
	}
}

func usage() {
	fmt.Println("Usage: diffnoise <image-path-or-url> <timestep> [output-path]")
	fmt.Println()
	fmt.Println("  image-path-or-url  local file path or http(s) URL")
	fmt.Println("  timestep           integer in [0, 999]")
	fmt.Println("  output-path        optional, defaults to noised_<t>.png")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  diffnoise photo.png 500")
	fmt.Println("  diffnoise photo.jpg 999 out.png")
	fmt.Println("  diffnoise https://example.com/img.png 250")
}

func compressImage(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	var cmd *exec.Cmd

	switch ext {
	case ".jpg", ".jpeg":
		// strip metadata + optimize huffman tables
		cmd = exec.Command("convert", path, "-strip", "-quality", "85", path)
	default:
		// png: max compression + strip metadata
		cmd = exec.Command("convert", path, "-strip", "-compress", "Zip", "-quality", "95", path)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("imagemagick: %v — %s", err, out)
	}
	return nil
}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	src := os.Args[1]
	t, err := strconv.Atoi(os.Args[2])
	if err != nil || t < 0 || t >= T {
		fmt.Fprintf(os.Stderr, "timestep must be an integer in [0, %d]\n", T-1)
		os.Exit(1)
	}

	outPath := fmt.Sprintf("noised_%d.png", t)
	if len(os.Args) >= 4 {
		outPath = os.Args[3]
	}

	// load image
	var img image.Image
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		fmt.Printf("Fetching %s...\n", src)
		img, err = loadImageFromURL(src)
	} else {
		img, err = loadImageFromDisk(src)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load image: %v\n", err)
		os.Exit(1)
	}

	alphaBar := buildSchedule()
	fmt.Printf("alpha_bar[%d] = %.6f\n", t, alphaBar[t])

	noised := noiseStep(img, t, alphaBar)

	if err := saveImage(noised, outPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save output: %v\n", err)
		os.Exit(1)
	}
	if err := compressImage(outPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: compression failed: %v\n", err)
		// non-fatal, image is still saved
	}
	fmt.Printf("saved -> %s\n", outPath)
}
