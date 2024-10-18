package datashape

import "gonum.org/v1/gonum/stat"

type (
	statsData []float64
)

func (s *statsData) add(v float64) {
	*s = append(*s, v)
}

func (s *statsData) Mean() float64 {
	return stat.Mean(*s, nil)
}

func (s *statsData) HarmonicMean() float64 {
	return stat.HarmonicMean(*s, nil)
}
