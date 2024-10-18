package sampler

// Stream sampling for variance-optimal estimation of subset sums
// Edith Cohen, Nick Duffield, Haim Kaplan, Carsten Lund, Mikkel Thorup
// 2008
// https://arxiv.org/pdf/0803.0473.pdf

import (
	"container/heap"
	"fmt"
	"math/rand"

	ms "github.com/jmacd/essay/examples/internal/multishape"
)

type (
	// Note: Using pointer-to-ms.Individual because the weights
	// are mutable during sampling.  TODO: Address the aliasing
	// problem somehow in multishape.

	varoptSampler struct {
		// Large-weight items
		L largeHeap

		// Light-weight items
		T []*ms.Individual

		// Temporary buffer
		X []*ms.Individual

		// Current threshold
		tau float64

		// Size of sample & scale
		K     int
		scale Scale

		totalCount  int
		totalWeight float64
	}

	largeHeap []*ms.Individual
)

func NewVaropt(size int, scale Scale) Sampler {
	return &varoptSampler{
		K:     size,
		scale: scale,
	}
}

func (s *varoptSampler) Add(individual ms.Individual) {
	ind := &individual

	wi := s.scale.Weigh(*ind)

	ind.SetWeight(wi)

	if wi <= 0 {
		panic(fmt.Sprint("Invalid weight <= 0: ", wi))
	}

	s.totalCount += 1
	s.totalWeight += wi

	if len(s.T)+len(s.L) < s.K {
		heap.Push(&s.L, ind)
		return
	}

	W := s.tau * float64(len(s.T))

	if wi > s.tau {
		heap.Push(&s.L, ind)
	} else {
		s.X = append(s.X, ind)
		W += wi
	}

	for len(s.L) > 0 && W >= float64(len(s.T)+len(s.X)-1)*s.L[0].Weight() {
		h := heap.Pop(&s.L).(*ms.Individual)
		s.X = append(s.X, h)
		W += h.Weight()
	}

	s.tau = W / float64(len(s.T)+len(s.X)-1)
	r := s.uniform()
	d := 0

	for d < len(s.X) && r >= 0 {
		wxd := s.X[d].Weight()
		r -= (1 - wxd/s.tau)
		d++
	}
	if r < 0 {
		if d < len(s.X) {
			s.X[d], s.X[len(s.X)-1] = s.X[len(s.X)-1], s.X[d]
		}
		s.X = s.X[:len(s.X)-1]
	} else {
		ti := rand.Intn(len(s.T))
		s.T[ti], s.T[len(s.T)-1] = s.T[len(s.T)-1], s.T[ti]
		s.T = s.T[:len(s.T)-1]
	}
	s.T = append(s.T, s.X...)
	s.X = s.X[:0]
}

func (s *varoptSampler) uniform() float64 {
	for {
		r := rand.Float64()
		if r != 0.0 {
			return r
		}
	}
}

// note is there a problem with aliasing arrays?

func addPop(o *ms.Population, p ms.Individual, w float64) {
	p.SetWeight(w)
	*o = append(*o, p)
}

func (s *varoptSampler) Weighted() ms.Population {
	var r ms.Population
	for _, p := range s.L {
		addPop(&r, *p, s.scale.EstimateWeight(*p, s.tau))
	}
	for _, p := range s.T {
		addPop(&r, *p, s.scale.EstimateWeight(*p, s.tau))
	}
	return r
}

func (s *varoptSampler) TotalWeight() float64 {
	return s.totalWeight
}

func (s *varoptSampler) TotalCount() int {
	return s.totalCount
}

func (s *varoptSampler) Tau() float64 {
	return s.tau
}

func (b largeHeap) Len() int {
	return len(b)
}

func (b largeHeap) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b largeHeap) Less(i, j int) bool {
	return b[i].Weight() < b[j].Weight()
}

func (b *largeHeap) Push(x interface{}) {
	*b = append(*b, x.(*ms.Individual))
}

func (b *largeHeap) Pop() interface{} {
	old := *b
	n := len(old)
	x := old[n-1]
	*b = old[0 : n-1]
	return x
}
