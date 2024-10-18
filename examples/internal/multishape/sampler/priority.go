package sampler

import (
	"math"

	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/wangjohn/quickselect"
)

type (
	prioritySampler struct {
		size   int
		buffer priorityBuffer
		scale  Scale

		totalWeight float64
		totalCount  int
	}

	priorityBuffer []ms.Individual
)

func NewPriority(size int, scale Scale) Sampler {
	return &prioritySampler{
		size:   size,
		scale:  scale,
		buffer: make(priorityBuffer, 0, (size+1)*2),
	}
}

func (s *prioritySampler) Add(ind ms.Individual) {
	// TODO note partial aliasing of ind.vars slices here, note
	// that multishape structs have fields dedicated to this
	// package, settle this confusion.
	w := s.scale.Weigh(ind)
	ind.SetWeight(w)
	s.totalWeight += w
	s.totalCount += 1

	s.compress()

	s.buffer = append(s.buffer, ind)
}

func (s *prioritySampler) TotalWeight() float64 {
	return s.totalWeight
}

func (s *prioritySampler) TotalCount() int {
	return s.totalCount
}

func (s *prioritySampler) Weighted() ms.Population {
	s.prioritize()
	if len(s.buffer) < s.size {
		return s.reweight(s.buffer)
	}
	return s.reweight(s.buffer[0:s.size])
}

func (s *prioritySampler) reweight(input []ms.Individual) ms.Population {
	copied := make([]ms.Individual, len(input))
	tau := s.Tau()
	for i := range input {
		copied[i] = input[i]
		copied[i].SetWeight(s.scale.EstimateWeight(copied[i], tau))
	}
	return copied
}

func (s *prioritySampler) Tau() float64 {
	if len(s.buffer) <= s.size {
		return math.NaN()
	}
	return s.buffer[s.size].Priority()
}

// compress makes room in the buffer by discarding elements whenever
// we reach twice the desired sample size. this achieves O(1)
// amortized insertion cost.
func (s *prioritySampler) compress() {
	if len(s.buffer) < cap(s.buffer) {
		return
	}
	s.prioritize()
}

// prioritize selects the maximum-priority s.size elements.
func (s *prioritySampler) prioritize() {
	if len(s.buffer) < s.size+1 {
		return
	}

	quickselect.QuickSelect(&s.buffer, s.size+1)
	s.buffer = s.buffer[0 : s.size+1]
}

func (b priorityBuffer) Len() int {
	return len(b)
}

func (b priorityBuffer) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b priorityBuffer) Less(i, j int) bool {
	return b[i].Priority() > b[j].Priority()
}
