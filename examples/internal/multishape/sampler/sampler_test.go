package sampler

import (
	"math"
	"math/rand"
	"testing"

	ms "github.com/jmacd/essay/examples/internal/multishape"
)

type (
	testScale struct{}
)

var (
	W ms.Variable = "W"
)

const (
	testLimit     = 100
	testSize      = 100
	testRepeats   = 100
	testTolerance = 0.02
)

func TestUnweightedPriority(t *testing.T) {
	testSampler(t, NewPriority, func(i int) float64 { return 1 })
}

func TestWeightedPriority(t *testing.T) {
	testSampler(t, NewPriority, func(i int) float64 { return float64(i) })
}

func TestUnweightedVaropt(t *testing.T) {
	testSampler(t, NewVaropt, func(i int) float64 { return 1 })
}

func TestWeightedVaropt(t *testing.T) {
	testSampler(t, NewVaropt, func(i int) float64 { return float64(i) })
}

func testSampler(t *testing.T, factory func(size int, scale Scale) Sampler, vf func(i int) float64) {
	var expect float64
	var sum float64

	for r := 0; r < testRepeats; r++ {
		s := factory(testSize, testScale{})

		for i := 1; i <= testLimit; i++ {
			for j := 1; j <= i; j++ {
				var ind ms.Individual

				w := vf(i)
				expect += w
				ind.AddNumerical(W, w, ms.CountMeasure)
				ind.SetAlpha(rand.Float64())

				s.Add(ind)
			}
		}

		if len(s.Weighted()) != testSize {
			t.Error("wrong size", len(s.Weighted()))
		}

		for _, ind := range s.Weighted() {
			sum += ind.Weight()
		}
	}

	if sum < expect*(1-testTolerance) || sum > expect*(1+testTolerance) {
		t.Error("sum mismatch", sum, expect)
	}
}

func (s testScale) Weigh(ind ms.Individual) float64 {
	return ind.Number(W)
}

func (s testScale) EstimateWeight(ind ms.Individual, tau float64) float64 {
	return math.Max(ind.Weight(), tau)
}
