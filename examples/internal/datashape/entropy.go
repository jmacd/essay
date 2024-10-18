package datashape

// See Entropy and the species accumulation curve: a novel entropy
// estimator via discovery rates of new species Anne Chao, Y.T. Wang,
// and Lou Jost.
//
// http://onlinelibrary.wiley.com/doi/10.1111/2041-210X.12108/epdf

import "math"

const (
	maxSampleSize = 100000
)

var (
	harmonics = make([]float64, maxSampleSize)
)

func init() {
	// A lookup table for harmonic numbers; note there is an approximation
	// formula, but it's a lot of calculation and we need potentially every
	// harmonic number from 1 to sampleSize-1
	for i := 1; i < maxSampleSize; i++ {
		harmonics[i] = harmonics[i-1] + 1/float64(i)
	}
}

func estimateEntropy(size, f1, f2 Frequency, frequencies Frequencies) float64 {
	if size > maxSampleSize {
		panic("Entropy estimator size is too large (need more harmonics)")
	}

	if len(frequencies) == 1 {
		return 0
	}
	h := 0.0
	dgn := harmonics[size-1]
	for _, xi := range frequencies {
		h += (dgn - harmonics[xi-1]) * float64(xi) / float64(size)
	}
	a := 0.0
	if f1 == 0 && f2 == 0 {
		a = 1
	} else if f2 == 0 {
		a = 2 / float64((size-1)*(f1-1)+2)
	} else {
		a = 2 * float64(f2) / float64((size-1)*f1+2*f2)
	}
	b := -math.Log(a)
	for r := Frequency(1); r < size-1; r++ {
		rf := float64(r)
		b -= math.Pow(1-a, rf) / rf
	}
	if a != 1 {
		h += float64(f1) * math.Pow(1-a, float64(1-size)) * b / float64(size)
	}
	return h
}
