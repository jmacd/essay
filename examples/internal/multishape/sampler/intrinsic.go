package sampler

import (
	"math"

	ms "github.com/jmacd/essay/examples/internal/multishape"
)

type (
	intrinsicScale struct {
		variable ms.Variable
	}
)

func Intrinsic(v ms.Variable) Scale {
	return intrinsicScale{v}
}

func (i intrinsicScale) Weigh(pt ms.Individual) float64 {
	return pt.Number(i.variable)
}

func (i intrinsicScale) EstimateWeight(pt ms.Individual, tau float64) float64 {
	return math.Max(pt.Number(i.variable), tau)
}
