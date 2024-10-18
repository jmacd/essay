package frequency

import (
	"math"
	"sort"

	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/examples/internal/multishape/sampler"
)

type (
	NumberPDF interface {
		ProbBucket(float64) float64
	}

	NumberScale struct {
		pdf    NumberPDF
		v      ms.Variable
		Points ms.Population
	}

	numberPDF struct {
		values []float64
		probs  []float64
	}

	neutralScale struct{}
)

func NewNumberScale(pdf NumberPDF, v ms.Variable, points ms.Population) sampler.Scale {
	return NumberScale{
		pdf:    pdf,
		v:      v,
		Points: points,
	}
}

func NewNumberPDF(values, probs []float64) NumberPDF {
	vcopy := make([]float64, len(values))
	pcopy := make([]float64, len(values))
	sum := 0.0
	for i := range values {
		vcopy[i] = values[i]
		sum += probs[i]
	}
	for i := range values {
		pcopy[i] = probs[i] / sum
	}
	return numberPDF{
		values: vcopy,
		probs:  pcopy,
	}
}

func (n numberPDF) ProbBucket(value float64) float64 {
	points := len(n.values) - 1

	idx := sort.Search(points, func(i int) bool {
		return n.values[i] >= value
	})
	if n.probs[idx] == 0 {
		panic("Invalid PDE")
	}
	return n.probs[idx]
}

func UniformScale() sampler.Scale {
	return neutralScale{}
}

func (fs NumberScale) Weigh(ind ms.Individual) float64 {
	return 1 / fs.pdf.ProbBucket(ind.Number(fs.v))
}

func (fs NumberScale) EstimateWeight(ind ms.Individual, tau float64) float64 {
	if math.IsNaN(tau) {
		return 1
	}
	weight := ind.Weight()
	pweight := math.Max(tau, weight)

	return pweight / weight
}

func (neutralScale) Weigh(ind ms.Individual) float64 {
	return 1
}

func (neutralScale) EstimateWeight(ind ms.Individual, tau float64) float64 {
	if math.IsNaN(tau) {
		return 1
	}
	return tau
}
