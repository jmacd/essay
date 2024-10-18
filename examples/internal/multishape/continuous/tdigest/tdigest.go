// tdigest uses the t-digest scaling function to estimate
// quantiles from a sample distribution.  What T-digest calls a
// "compression" paramter is really a quality parameter.  As presented
// here, it exactly determines the number of quantiles.
//
// https://github.com/tdunning/t-digest/blob/7aaa5b76e7e590967277dac45baa34c12e0b717b/docs/t-digest-paper/histo.pdf
package tdigest

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/jmacd/essay/examples/internal/multishape/continuous"
)

const pi = math.Pi

type (
	// Quality is a quality parameter, must be > 0.
	//
	// Forcing the compression parameter to be an integer
	// simplifies the code, because each centroid / quantile has
	// k(i) == 1.
	Quality int

	// input is for sorting.
	input struct {
		values  []float64
		weights []float64
	}

	// TDigest is computed using Algorithm 1 from the T-digest
	// paper.
	TDigest struct {
		Quality Quality
		digest  []continuous.Centroid
		VRange  continuous.Range
		continuous.Boundary
	}
)

// The T-digest scaling function maps quantile to index, where
// 0 <= index < quality
func (quality Quality) digestK(q float64) float64 {
	return float64(quality) * (0.5 + math.Asin(2*q-1)/pi)
}

// The T-digest inverse-scaling function maps index to quantile.
func (quality Quality) inverseK(k float64) float64 {
	if k > float64(quality) {
		return 1
	}
	return 0.5 * (math.Sin(pi*(k/float64(quality)-0.5)) + 1)
}

// TODO: this algorithm treats the first and last centroid
// differently--one centroid will be smaller than the other--which may
// show up as a different kinds of bias depending on how the data are
// used.  This code should provide an option to choose where the bias
// is more or less important, it just means changing the data sort order
// and reversing the centroids afterwards.
func (quality Quality) New(origValues, origWeights []float64, vrange continuous.Range, boundary continuous.Boundary) continuous.Digest {
	if len(origValues) == 0 || len(origWeights) != len(origValues) {
		panic("Invalid input")
	}
	in := input{
		values:  make([]float64, len(origValues)),
		weights: make([]float64, len(origValues)),
	}

	copy(in.values, origValues)
	copy(in.weights, origWeights)

	sumWeight := 0.0
	for _, we := range in.weights {
		if we > 100e100 {
			panic(fmt.Sprint("Insane weight", we))
		}
		sumWeight += we
	}

	min := math.Inf(+1)
	max := math.Inf(-1)
	for _, va := range in.values {
		if va < vrange.Min {
			va = vrange.Min
		} else if va > vrange.Max {
			va = vrange.Max
		}
		min = math.Min(va, min)
		max = math.Max(va, max)
	}

	sort.Sort(in)

	var digest []continuous.Centroid

	qleft := 0.0
	qlimit := quality.inverseK(1)

	current := continuous.Centroid{
		Sum:    in.values[0] * in.weights[0],
		Weight: in.weights[0],
		Count:  1,
	}

	for pos := 1; pos < len(in.values); pos++ {
		vpos := in.values[pos]
		wpos := in.weights[pos]

		q := qleft + (current.Weight+wpos)/sumWeight

		if q <= qlimit {
			current.Sum += vpos * wpos
			current.Weight += wpos
			current.Count++
			continue
		}

		digest = append(digest, current)
		qleft += current.Weight / sumWeight

		qlimit = quality.inverseK(quality.digestK(qleft) + 1)
		current.Weight = wpos
		current.Sum = vpos * wpos
		current.Count = 1
	}
	digest = append(digest, current)

	sum := 0.0
	size := len(digest)
	for ki := 0; ki < size-1; ki++ {
		sum += digest[ki].Weight
		digest[ki].UpperValue = 0.5 * (digest[ki].Mean() + digest[ki+1].Mean())
		digest[ki].UpperQuantile = sum / sumWeight
	}
	digest[size-1].UpperValue = max
	digest[size-1].UpperQuantile = 1

	for ki := size - 1; ki > 0; ki-- {
		digest[ki].LowerValue = digest[ki-1].UpperValue
	}
	digest[0].LowerValue = min

	lower := 0.0
	for i := range digest {
		part := (digest[i].UpperQuantile - lower) / (digest[i].UpperValue - digest[i].LowerValue)

		if part < 0 {
			panic(fmt.Sprintln("Part", i, "is negative"))
		}

		digest[i].ProbDensity = part
		lower = digest[i].UpperQuantile
	}

	return &TDigest{
		Quality:  quality,
		digest:   digest,
		VRange:   vrange,
		Boundary: boundary,
	}
}

func (s *TDigest) Summary() []continuous.Centroid {
	return append(s.digest[:0:0], s.digest...)
}

func (s *TDigest) Size() int {
	return len(s.digest)
}

// ProbDensity returns the probability mass function.
func (s *TDigest) ProbDensity(value float64) float64 {
	idx := s.Lookup(value)
	centroid := s.digest[idx]

	if idx == 0 {
		return s.boundaryProbDensity(value, centroid, -1, s.VRange.Min)
	}
	if idx == len(s.digest)-1 {
		return s.boundaryProbDensity(value, centroid, +1, s.VRange.Max)
	}
	return centroid.ProbDensity
}

// ProbBucket returns the average probability mass function of the nearest centroid.
func (s *TDigest) ProbBucket(value float64) float64 {
	idx := s.Lookup(value)
	return s.digest[idx].Prob()
}

func (s *TDigest) Lookup(value float64) int {
	points := len(s.digest) - 1

	idx := sort.Search(points, func(i int) bool {
		return s.digest[i].UpperValue >= value
	})
	return idx
}

func (s *TDigest) boundaryProbDensity(input float64, bin continuous.Centroid, sign int, limit float64) float64 {
	// Return 0 probability for out-of-range values.
	if math.IsNaN(input) {
		return 0
	}

	leftside := sign < 0
	extreme := false
	unsupported := false

	if leftside {
		unsupported = input < limit
		extreme = input < bin.LowerValue
	} else {
		unsupported = input > limit
		extreme = input > bin.UpperValue
	}

	if unsupported {
		// Return 0 probability for unsupported values.
		return 0
	}

	if s.Boundary == continuous.Extremal {
		// Return 0 probability for below-min or above-max.
		if extreme {
			return 0
		}
		return bin.ProbDensity
	}

	// Offset is the (positive) distance to the non-extreme end of the bin.
	offset := 0.0

	if leftside {
		offset = bin.UpperValue - input
	} else {
		offset = input - bin.LowerValue
	}

	// In both cases below, place half of the probability in the
	// bin, half in the remaining area.
	delta0 := bin.UpperValue - bin.LowerValue
	area := bin.ProbDensity * delta0

	// For infinite limit, use a negative exponential distribution.
	if math.IsInf(limit, sign) {
		lambda := math.Ln2 / delta0
		prob := lambda * math.Exp(-lambda*offset)
		density := area * prob
		if math.IsInf(density, 0) {
			panic("Inf density")
		}
		return density
	}

	// The distribution is truncated, divide evenly into a
	// right trapezoid (width==delta0) and a rectangle (width=delta1).
	delta1 := 0.0

	if leftside {
		delta1 = bin.LowerValue - limit
	} else {
		delta1 = limit - bin.UpperValue
	}

	// If the limit was met, just do it.
	if delta1 == 0 {
		return bin.ProbDensity
	}

	// If the remaining region is smaller than the bin width, a flat reponse.
	if delta1 < delta0 {
		return area / (delta0 + delta1)
	}

	p1 := 0.5 * area / delta1

	// Is the input outside of the extreme bin?
	if offset > delta0 {
		return p1
	}

	p0 := area/delta0 - p1
	return p1 + (1-(offset/delta0))*(p0-p1)
}

func (s *TDigest) Quantiles(f func(ctrd continuous.Centroid)) {
	for _, ctrd := range s.digest {
		f(ctrd)
	}
}

func (in input) Len() int {
	return len(in.values)
}

func (in input) Swap(i, j int) {
	in.values[i], in.values[j] = in.values[j], in.values[i]
	in.weights[i], in.weights[j] = in.weights[j], in.weights[i]
}

func (in input) Less(i, j int) bool {
	return in.values[i] < in.values[j]
}

func (s *TDigest) String() string {
	var b strings.Builder
	p := 0.0
	for _, x := range s.digest {
		p += x.Prob()
		b.WriteString(fmt.Sprintf("mass=%.2f mean=%.2e count=%d prob=%.2e\n", x.Weight, x.Mean(), x.Count, x.Prob()))
	}
	b.WriteString(fmt.Sprint("P total: ", p))
	return b.String()
}

func (s *TDigest) DataRange() (xmin, xmax, ymin, ymax float64) {
	for _, c := range s.digest {
		ymax = math.Max(ymax, c.ProbDensity)
	}
	xmin = math.Min(xmin, s.digest[0].LowerValue)
	xmax = math.Max(xmax, s.digest[len(s.digest)-1].UpperValue)
	return
}
