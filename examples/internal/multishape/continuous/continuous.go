package continuous

import "math"

type (
	Centroid struct {
		Sum           float64
		Weight        float64
		LowerValue    float64
		UpperValue    float64
		UpperQuantile float64
		ProbDensity   float64
		Count         int
	}

	Range struct {
		Min, Max float64
	}

	// Boundary determines how the extremal buckets are sized.
	Boundary int

	Digest interface {
		Size() int
		Lookup(value float64) int
		Summary() []Centroid
		ProbDensity(value float64) float64
		ProbBucket(value float64) float64
		DataRange() (xmin, xmax, ymin, ymax float64)
	}

	Builder interface {
		New(origValues, origWeights []float64, vrange Range, boundary Boundary) Digest
	}
)

const (
	Extremal = iota
	RangeSupported
)

func FullRange() Range {
	return Range{
		Min: math.Inf(-1),
		Max: math.Inf(+1),
	}
}

func PositiveRange() Range {
	return Range{
		Min: 0,
		Max: math.Inf(+1),
	}
}

func (c Centroid) Prob() float64 {
	return c.ProbDensity * (c.UpperValue - c.LowerValue)
}

func (c Centroid) Mean() float64 {
	return c.Sum / c.Weight
}
