package sampler

import (
	ms "github.com/jmacd/essay/examples/internal/multishape"
)

type (
	Sampler interface {
		Add(ind ms.Individual)
		TotalWeight() float64
		TotalCount() int
		Weighted() ms.Population
		Tau() float64
	}

	Model interface {
		Update(sampled ms.Population, prior Scale) Scale
	}

	Scale interface {
		// Weigh an item.
		Weigh(ms.Individual) float64

		// Estimate the population total for an item.  `tau`
		// is the priority sampling threshold.
		EstimateWeight(item ms.Individual, tau float64) float64
	}
)

func Compute(wholePop ms.Population, prior Scale, size int, model Model) (ms.Population, Scale) {
	// samp := NewPriority(size, prior)
	samp := NewVaropt(size, prior)

	for _, ind := range wholePop {
		samp.Add(ind)
	}

	sampled := samp.Weighted()
	posterior := model.Update(sampled, prior)
	return sampled, posterior
}
