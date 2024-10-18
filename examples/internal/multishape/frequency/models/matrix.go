package models

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

func constMat(r, c int, v float64) *mat.Dense {
	m := mat.NewDense(r, c, nil)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			m.Set(i, j, v)
		}
	}
	return m
}

func diagonalize(v mat.Vector) mat.Matrix {
	d := mat.NewDiagDense(v.Len(), nil)
	for i := 0; i < v.Len(); i++ {
		d.SetDiag(i, v.AtVec(i))
	}
	return d
}

func invSqrt(_, _ int, value float64) float64 {
	if value == 0 {
		return 0
	}
	return 1 / math.Sqrt(value)
}

func sqrt(_, _ int, value float64) float64 {
	if value == 0 {
		return 0
	}
	return math.Sqrt(value)
}

func inverse(_, _ int, value float64) float64 {
	if value == 0 {
		return 0
	}
	return 1 / value
}
