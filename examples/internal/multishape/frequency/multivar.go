package frequency

import (
	"math"

	ms "github.com/jmacd/essay/examples/internal/multishape"
)

type (
	MultiVarScale struct {
		// Weight functions
		wfs []func(ms.Individual) float64

		// A copy of the former sampled data
		Sampled ms.Population

		// TODO this is temporary b.s.
		Knowns      int
		EstUnknowns float64
		UnknownProb float64
		TotalWeight float64
		Map         map[ms.Category]int
		Probs       map[ms.Category]float64
	}
)

func NewMultiVarScale(sampled ms.Population) *MultiVarScale {
	return &MultiVarScale{Sampled: sampled}
}

func (mvs *MultiVarScale) Add(f func(ms.Individual) float64) *MultiVarScale {
	mvs.wfs = append(mvs.wfs, f)
	return mvs
}

func (mvs *MultiVarScale) Weigh(p ms.Individual) float64 {
	sum := 0.0
	for _, f := range mvs.wfs {
		sum += f(p)
	}
	return math.Sqrt(sum)
}

func (mvs *MultiVarScale) EstimateWeight(item ms.Individual, tau float64) float64 {
	if math.IsNaN(tau) {
		return 1
	}
	weight := item.Weight()
	pweight := math.Max(tau, weight)

	return pweight / weight
}
