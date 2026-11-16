package shapes

import "image/color"

// ShapeClass describes a polygon class used in the dataset.
type ShapeClass struct {
	ID    int
	Name  string
	Sides int
}

// Classes is the full list of supported polygon classes.
var Classes = []ShapeClass{
	{0, "triangle", 3},
	{1, "quadrilateral", 4},
	{2, "pentagon", 5},
	{3, "hexagon", 6},
	{4, "heptagon", 7},
	{5, "octagon", 8},
	{6, "decagon", 10},
}

// FillColors is the palette used for polygon fill and stroke colors.
var FillColors = []color.NRGBA{
	{R: 230, G: 57, B: 70, A: 255},
	{R: 244, G: 162, B: 97, A: 255},
	{R: 42, G: 157, B: 143, A: 255},
	{R: 69, G: 123, B: 157, A: 255},
	{R: 155, G: 93, B: 229, A: 255},
	{R: 247, G: 37, B: 133, A: 255},
	{R: 76, G: 201, B: 240, A: 255},
	{R: 128, G: 185, B: 24, A: 255},
	{R: 255, G: 107, B: 53, A: 255},
	{R: 168, G: 218, B: 220, A: 255},
}

// BgColors is the palette used for image backgrounds.
var BgColors = []color.NRGBA{
	{R: 255, G: 255, B: 255, A: 255},
	{R: 248, G: 249, B: 250, A: 255},
	{R: 26, G: 26, B: 46, A: 255},
	{R: 15, G: 52, B: 96, A: 255},
	{R: 22, G: 33, B: 62, A: 255},
	{R: 245, G: 245, B: 240, A: 255},
	{R: 240, G: 237, B: 255, A: 255},
	{R: 20, G: 30, B: 20, A: 255},
}
