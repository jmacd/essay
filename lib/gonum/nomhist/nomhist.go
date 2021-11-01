// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nomhist

import (
	"image/color"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// A NominalHistogram presents grouped data with rectangular bars
// with lengths proportional to the data values.
type NominalHistogram struct {
	plotter.Values

	// Color is the fill color of the bars.
	Color func(int) color.Color

	// LineStyle is the style of the outline of the bars.
	draw.LineStyle
}

// NewNominalHistogram returns a new bar chart with a single bar for each value.
// The bars heights correspond to the values and their x locations correspond
// to the index of their value in the Valuer.
func NewNominalHistogram(vs plotter.Valuer, cf func(int) color.Color) (*NominalHistogram, error) {
	values, err := plotter.CopyValues(vs)
	if err != nil {
		return nil, err
	}
	return &NominalHistogram{
		Values:    values,
		Color:     cf,
		LineStyle: plotter.DefaultLineStyle,
	}, nil
}

// Plot implements the plot.Plotter interface.
func (b *NominalHistogram) Plot(c draw.Canvas, plt *plot.Plot) {
	trCat, trVal := plt.Transforms(&c)

	width := c.Size().X / vg.Length(1+len(b.Values))

	for i, ht := range b.Values {
		catVal := float64(i)
		catMin := trCat(float64(catVal))
		if !c.ContainsX(catMin) {
			continue
		}
		catMin = catMin - width/2
		catMax := catMin + width
		bottom := 0.0
		valMin := trVal(bottom)
		valMax := trVal(bottom + ht)

		pts := []vg.Point{
			{catMin, valMin},
			{catMin, valMax},
			{catMax, valMax},
			{catMax, valMin},
		}
		poly := c.ClipPolygonY(pts)
		c.FillPolygon(b.Color(i), poly)

		pts = append(pts, vg.Point{X: catMin, Y: valMin})
		outline := c.ClipLinesY(pts)
		c.StrokeLines(b.LineStyle, outline...)
	}
}

// DataRange implements the plot.DataRanger interface.
func (b *NominalHistogram) DataRange() (xmin, xmax, ymin, ymax float64) {
	catMin := -1.
	catMax := float64(len(b.Values))

	valMin := math.Inf(1)
	valMax := math.Inf(-1)
	for _, val := range b.Values {
		valBot := 0.0
		valTop := valBot + val
		valMin = math.Min(valMin, math.Min(valBot, valTop))
		valMax = math.Max(valMax, math.Max(valBot, valTop))
	}
	return catMin, catMax, valMin, valMax
}
