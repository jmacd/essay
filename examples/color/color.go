package main

import (
	"fmt"
	"image"
	"math"

	"github.com/jmacd/essay"
	colorful "github.com/lucasb-eyer/go-colorful"
)

type (
	colors struct{}
)

func main() {
	essay.Main("Perceptual Color Spaces", write)
}

func write(doc essay.Document) {
	doc.Note(
		`This essay demonstrates the features of the essay
		library using the go-colorful library.`,

		`There are many available color spaces we may use to
		represent color in print and on displays.  We may be
		interested in choosing a palette of colors to display
		categorical information on a scatter plot, or we may
		be using color gradients to represent continuous
		variables.`,

		`There are "perceptually" scaled color spaces,
		determined on the basis of human studies, chosen so
		that humans will perceive uniform variation across the
		gradient.  That is our eyes pereieve the "distance"
		between colors proportional to their numerical
		distance.  Non-colorblind human viewers are expected
		to see an uniform color gradient on any line cut
		through the two-dimensional color plane.`,

		`Below, a table of three perceptual color spaces,
		presented as a square (rectangular coordinates) or a
		wheel (polar coordinates), for intensity levels from
		10% to 100%.`)

	doc.Section("Color tiles", &colors{})
	doc.Section("Blending", blending)
}

func (p *colors) Display(doc essay.Document) {
	var cells [][]interface{}
	var left []interface{}
	top := []interface{}{
		"Level",
		"LAB",
		"LUV",
		"LCh",
	}

	for level := 0.1; level <= 1; level += .1 {
		lc := p.displayLevel(level)
		left = append(left, fmt.Sprintf("%.0f%%", level*100))
		cells = append(cells, lc)
	}

	doc.Note(essay.Table{
		Cells:   cells,
		TopRow:  top,
		LeftCol: left,
	})
}

func (c *colors) displayLevel(level float64) []interface{} {
	const (
		h = 300
		w = 300
		π = math.Pi
	)

	rmax := math.Sqrt(h * w / π)

	imgLAB := image.NewRGBA(image.Rect(0, 0, w, h))
	imgLUV := image.NewRGBA(image.Rect(0, 0, w, h))
	imgLCh := image.NewRGBA(image.Rect(0, 0, int(rmax*2), int(rmax*2)))

	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			a := 2 * float64(x-w/2) / w
			b := 2 * float64(y-h/2) / h
			imgLAB.Set(x, y, colorful.Lab(level, a, b).Clamped())
			imgLUV.Set(x, y, colorful.Luv(level, a, b).Clamped())
		}
	}

	for r := 0.0; r < rmax; r++ {
		const (
			π180   = π / 180
			circle = 2 * π
			step   = circle / 5000
		)
		for θ := 0.0; θ < circle; θ += step {
			imgLCh.Set(
				int(rmax+math.Cos(θ)*r),
				int(rmax+math.Sin(θ)*r),
				colorful.Hcl(
					θ/π180,
					r/rmax,
					level,
				).Clamped(),
			)
		}
	}

	return []interface{}{
		essay.Image(imgLAB),
		essay.Image(imgLUV),
		essay.Image(imgLCh),
	}
}

func blending(doc essay.Document) {
	red := colorful.Color{R: 255}
	blue := colorful.Color{B: 255}
	green := colorful.Color{G: 255}
	yellow := colorful.Color{R: 255, G: 255}

	doc.Note("One", essay.Animation(
		essay.Image(blend(red, blue)),
		essay.Image(blend(green, yellow)),
		essay.Image(blend(blue, red)),
		essay.Image(blend(yellow, green)),
	))

	doc.Note("Two", essay.Animation(
		essay.Image(blend(red, yellow)),
		essay.Image(blend(green, blue)),
		essay.Image(blend(yellow, red)),
		essay.Image(blend(blue, green)),
	))

	doc.Note("Three", essay.Animation(
		essay.Image(blend(red, green)),
		essay.Image(blend(yellow, blue)),
		essay.Image(blend(green, red)),
		essay.Image(blend(blue, yellow)),
	))
	doc.Note("Four", essay.Animation(
		essay.Image(blend(red, yellow)),
		essay.Image(blend(yellow, blue)),
		essay.Image(blend(blue, green)),
		essay.Image(blend(green, red)),
	))
}

func blend(col1, col2 colorful.Color) image.Image {
	w := 500
	h := 500
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// TODO: colorful's Blend* functions don't work well at the
	// edges, needs internal clamping.  This shows it.
	for i := 0; i < w; i++ {
		c := col1.BlendHcl(col2, float64(i)/float64(w)).Clamped()
		for j := 0; j < h; j++ {
			img.Set(i, j, c)
		}
	}

	return img
}
