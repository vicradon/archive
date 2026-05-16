package render

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"

	"shapegen/shapes"
)

// Vec2 is a 2D point.
type Vec2 struct{ X, Y float64 }

// BBox is an axis-aligned bounding box.
type BBox struct{ X, Y, W, H float64 }

// GeneratedImage holds the encoded PNG bytes and the shape bounding box.
type GeneratedImage struct {
	Data []byte
	BBox BBox
}

// ClassImages groups all generated images for a single shape class.
type ClassImages struct {
	Class  shapes.ShapeClass
	Images []GeneratedImage
}

func polygonVertices(cx, cy, r float64, sides int, rotation float64) []Vec2 {
	pts := make([]Vec2, sides)
	for i := 0; i < sides; i++ {
		angle := 2*math.Pi*float64(i)/float64(sides) + rotation
		pts[i] = Vec2{cx + r*math.Cos(angle), cy + r*math.Sin(angle)}
	}
	return pts
}

// Bresenham-style filled polygon using scanline.
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
	img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
}

// Draw anti-aliased line using Wu's algorithm.
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

// GenerateShapeImage renders a random polygon on a random background and
// returns the NRGBA image together with its bounding box.
func GenerateShapeImage(rng *rand.Rand, size, sides int) (*image.NRGBA, BBox) {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	// Background
	bg := shapes.BgColors[rng.Intn(len(shapes.BgColors))]
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	// Optional background noise
	if rng.Float64() > 0.6 {
		for i := 0; i < 25; i++ {
			nc := shapes.FillColors[rng.Intn(len(shapes.FillColors))]
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

	fillIdx := rng.Intn(len(shapes.FillColors))
	fc := shapes.FillColors[fillIdx]
	fillAlpha := 0.55 + rng.Float64()*0.4

	pts := polygonVertices(cx, cy, r, sides, rotation)
	fillPolygon(img, pts, fc, fillAlpha)

	// Stroke — pick a different color
	strokeIdx := (fillIdx + 1 + rng.Intn(len(shapes.FillColors)-1)) % len(shapes.FillColors)
	sc := shapes.FillColors[strokeIdx]
	lw := math.Max(1.0, float64(size)*0.025)
	strokePolygon(img, pts, sc, lw)

	// Compute bounding box
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, p := range pts {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	minX = math.Max(0, minX)
	minY = math.Max(0, minY)
	maxX = math.Min(float64(size), maxX)
	maxY = math.Min(float64(size), maxY)

	return img, BBox{minX, minY, maxX - minX, maxY - minY}
}
