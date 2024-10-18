// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loghist

import (
	"fmt"
	"image/color"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type (
	// Histogram implements the Plotter interface,
	// drawing a histogram of the data.
	Histogram struct {
		// Bins is the set of bins for this histogram.
		Bins []HistogramBin

		// FillColor is the color used to fill each
		// bar of the histogram.  If the color is nil
		// then the bars are not filled.
		FillColors []color.Color

		// May be set to LinearTransformer or LogTransformer.
		Transform DataTransformer

		// LineStyle is the style of the outline of each
		// bar of the histogram.
		draw.LineStyle

		// If true, bars are placed at their centroid, not strictly
		// adjacent to each other.
		centroids bool
		normalize float64

		TotalWidth float64
		SumWeight  float64
	}

	DataTransformer interface {
		Transform(float64) float64
		Invert(float64) float64
	}

	Binner interface {
		BinPoints(xys plotter.XYer, transform DataTransformer) []HistogramBin
	}

	// A HistogramBin approximates the number of values
	// within a range by a single number (the weight).
	HistogramBin struct {
		Min, Max float64
		Weight   float64
		Sum      float64
	}

	LinearTransformer struct{}

	LogTransformer struct{}

	FixedBinner struct {
		Count int
		Min   float64
		Max   float64
	}

	LinearBinner struct {
		Count int
	}
)

func NewFixedBinner(count int, min, max float64) FixedBinner {
	return FixedBinner{
		Count: count,
		Min:   min,
		Max:   max,
	}
}

func NewLinearBinner(count int) LinearBinner {
	return LinearBinner{
		Count: count,
	}
}

// NewHistogram returns a new histogram
// that represents the distribution of values
// using the given number of bins.
//
// Each y value is assumed to be the frequency
// count for the corresponding x.
//
// If the number of bins is non-positive than
// a reasonable default is used.
func NewHistogram(transform DataTransformer, bins []HistogramBin) *Histogram {
	weight := 0.0
	width := bins[len(bins)-1].Max - bins[0].Min
	for _, b := range bins {
		weight += b.Weight
	}
	return &Histogram{
		Bins:       bins,
		Transform:  transform,
		FillColors: nil,
		LineStyle:  plotter.DefaultLineStyle,
		centroids:  false,
		SumWeight:  weight,
		TotalWidth: width,
		normalize:  weight,
	}
}

func ValuerToBins(vs plotter.Valuer, transform DataTransformer, binner Binner) []HistogramBin {
	return binner.BinPoints(UnitYs{vs}, transform)
}

type UnitYs struct {
	plotter.Valuer
}

func (u UnitYs) XY(i int) (float64, float64) {
	return u.Value(i), 1.0
}

func (u UnitYs) XYok(i int) (float64, float64, bool) {
	return u.Value(i), 1.0, true
}

// Plot implements the Plotter interface, drawing a line
// that connects each point in the Line.
func (h *Histogram) Plot(c draw.Canvas, p *plot.Plot) {
	trX, trY := p.Transforms(&c)

	for idx, bin := range h.Bins {
		min, max, height := h.coords(bin)
		pts := []vg.Point{
			{trX(min), trY(0)},
			{trX(max), trY(0)},
			{trX(max), trY(height)},
			{trX(min), trY(height)},
		}
		if h.FillColors != nil {
			c.FillPolygon(h.FillColors[idx], c.ClipPolygonXY(pts))
		}
		pts = append(pts, pts[0])
		c.StrokeLines(h.LineStyle, c.ClipLinesXY(pts)...)
	}
}

func (h *Histogram) coords(bin HistogramBin) (min, max, height float64) {
	var center float64
	width := bin.Max - bin.Min
	half := width / 2
	if h.centroids {
		center = bin.Sum / bin.Weight
	} else {
		center = bin.Min + half
	}
	//height = (bin.Weight / h.SumWeight) / width
	// @@@ WTF * h.normalize
	height = bin.Weight
	return center - half, center + half, height
}

// DataRange returns the minimum and maximum X and Y values
func (h *Histogram) DataRange() (xmin, xmax, ymin, ymax float64) {
	xmin = math.Inf(1)
	xmax = math.Inf(-1)
	ymax = math.Inf(-1)
	for _, bin := range h.Bins {
		min, max, height := h.coords(bin)
		if max > xmax {
			xmax = max
		}
		if min < xmin {
			xmin = min
		}
		if height > ymax {
			ymax = height
		}
	}
	return
}

func (h *Histogram) Centroids() {
	h.centroids = true
}

// Normalize normalizes the histogram so that the
// total area beneath it sums to a given value.
func (h *Histogram) Normalize(to float64) {
	h.normalize = to
}

// Thumbnail draws a rectangle in the given style of the histogram.
func (h *Histogram) Thumbnail(c *draw.Canvas) {
	ymin := c.Min.Y
	ymax := c.Max.Y
	xmin := c.Min.X
	xmax := c.Max.X

	pts := []vg.Point{
		{xmin, ymin},
		{xmax, ymin},
		{xmax, ymax},
		{xmin, ymax},
	}
	if h.FillColors != nil {
		c.FillPolygon(h.FillColors[0], c.ClipPolygonXY(pts))
	}
	pts = append(pts, pts[0])
	c.StrokeLines(h.LineStyle, c.ClipLinesXY(pts)...)
}

// binPoints returns a slice containing the given number of bins.
//
// If the given number of bins is not positive
// then a reasonable default is used.  The
// default is the square root of the sum of
// the y values.
func (f FixedBinner) binPoints(xys plotter.XYer, transform DataTransformer) []HistogramBin {
	n := f.Count
	if n <= 0 {
		m := 0.0
		for i := 0; i < xys.Len(); i++ {
			_, y := xys.XY(i)
			m += math.Max(y, 1.0)
		}
		n = int(math.Ceil(math.Sqrt(m)))
	}
	if n < 1 || f.Max <= f.Min {
		n = 1
	}

	bins := make([]HistogramBin, n)

	minXlog := transform.Transform(f.Min)
	w := (transform.Transform(f.Max) - minXlog) / float64(n)
	for i := range bins {
		bins[i].Min = transform.Invert(minXlog + float64(i)*w)
		bins[i].Max = transform.Invert(minXlog + float64(i+1)*w)
	}

	for i := 0; i < xys.Len(); i++ {
		x, y := xys.XY(i)
		bin := int((transform.Transform(x) - minXlog) / w)
		if x == f.Max {
			bin = n - 1
		}
		if bin < 0 || bin >= n {
			panic(fmt.Sprintf("%g, xmin=%g, xmax=%g, w=%g, bin=%d, n=%d\n",
				x, f.Min, f.Max, w, bin, n))
		}
		bins[bin].Weight += y
		bins[bin].Sum += x * y
	}
	return bins
}

func (l LinearTransformer) Transform(f float64) float64 { return f }
func (l LinearTransformer) Invert(g float64) float64    { return g }

func (l LogTransformer) Transform(f float64) float64 {
	if f <= 0 {
		panic("Values must be greater than 0 for a log scale")
	}
	return math.Log(f)
}
func (l LogTransformer) Invert(g float64) float64 { return math.Exp(g) }

func (f FixedBinner) BinPoints(xys plotter.XYer, transform DataTransformer) []HistogramBin {
	return f.binPoints(xys, transform)
}

func (b LinearBinner) BinPoints(xys plotter.XYer, transform DataTransformer) []HistogramBin {
	avg := float64(xys.Len()) / float64(b.Count)

	var bins []HistogramBin
	var idx int
	for i := 1; i <= b.Count; i++ {
		bin := HistogramBin{
			Min: math.Inf(+1),
			Max: math.Inf(-1),
		}
		for idx < int(math.Round(float64(i)*avg)) {
			v, w := xys.XY(idx)
			bin.Min = min(bin.Min, v)
			bin.Max = max(bin.Max, v)
			bin.Weight += w
			bin.Sum += v
			idx++
		}
		bins = append(bins, bin)
	}
	return bins
}
