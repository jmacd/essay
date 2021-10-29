package num

import (
	"math"

	"gonum.org/v1/plot/plotter"
)

type (
	XYs struct {
		X, Y []float64
	}
)

func (xys *XYs) Add(value, weight float64) {
	xys.X = append(xys.X, value)
	xys.Y = append(xys.Y, weight)
}

func (xys XYs) Len() int {
	return len(xys.X)
}

func (xys XYs) XY(i int) (x, y float64) {
	return xys.X[i], xys.Y[i]
}

func (xys XYs) XYok(i int) (x, y float64, ok bool) {
	return xys.X[i], xys.Y[i], true
}

func (xys XYs) MaxY() float64 {
	m := 0.0
	for i := 0; i < xys.Len(); i++ {
		m = math.Max(m, xys.Y[i])
	}
	return m
}

func Filter(xys plotter.XYer, f func(x, y float64) bool) XYs {
	filtered := XYs{}
	for i := 0; i < xys.Len(); i++ {
		x, y := xys.XY(i)
		if !f(x, y) {
			continue
		}
		filtered.X = append(filtered.X, x)
		filtered.Y = append(filtered.Y, y)
	}
	return filtered
}
