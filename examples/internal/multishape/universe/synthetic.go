package universe

import (
	"fmt"
	"math"
	"sort"

	ms "github.com/jmacd/essay/examples/internal/multishape"
	"golang.org/x/exp/rand"
)

// TODO this code is THE WORST.

const (
	// A "universal" variable.
	Time ms.Variable = "time"

	// Initial time value.
	StartTime = 0.0
)

var (
	NotSupported error = fmt.Errorf("not supported")
)

type (
	Universe struct {
		Random  *rand.Rand
		name    string
		seed    uint64
		nextvar uint64
	}

	MultiVariate interface {
		Produce(vs *ms.Individual) error
	}

	SingleVariate interface {
		Generate(name ms.Variable, vs *ms.Individual) error
	}

	NamedVariate struct {
		Name ms.Variable
		V    SingleVariate
	}

	AnonRander interface {
		Rand() float64
		Prob(float64) float64
	}

	Rander interface {
		AnonRander
		Name() string
	}

	InRange struct {
		value SingleVariate

		min, max float64
	}

	DiscreteDistribution []float64

	Categorical struct {
		dist   DiscreteDistribution // a CDF
		values []ms.Category
		rnd    *rand.Rand
	}

	Numerical struct {
		rnd Rander
	}

	Poisson struct {
		rate  float64
		value float64
		end   float64
		rnd   *rand.Rand
	}

	LeafProfile struct {
		*Universe
		names    []ms.Variable
		variates []SingleVariate
	}

	MixedProfile struct {
		*Universe
		dist     DiscreteDistribution
		profiles []MultiVariate
	}
)

func New(name string, seed uint64) *Universe {
	return &Universe{
		Random:  rand.New(rand.NewSource(seed)),
		name:    name,
		seed:    seed,
		nextvar: 1,
	}
}

func (u *Universe) Produce(v MultiVariate) ms.Population {
	var points ms.Population
	for {
		var pt ms.Individual
		if err := v.Produce(&pt); err != nil {
			if err != NotSupported {
				panic(fmt.Sprint("Unsupported error in Produce:", err))
			}
			break
		}
		points = append(points, pt)
	}
	return points
}

func (u *Universe) NewLeafProfile() *LeafProfile {
	return &LeafProfile{
		Universe: u,
	}
}

func (u *Universe) NewMixedProfile(dist Rander, profiles ...MultiVariate) *MixedProfile {
	return &MixedProfile{
		Universe: u,
		dist:     toDiscreteDistribution(dist, len(profiles)),
		profiles: profiles,
	}
}

func (u *Universe) NewCategorical(dist Rander, values []ms.Category) SingleVariate {
	return &Categorical{
		values: values,
		dist:   toDiscreteDistribution(dist, len(values)),
		rnd:    u.Random,
	}
}

func (u *Universe) InRange(v SingleVariate, min, max float64) SingleVariate {
	return &InRange{
		value: v,
		min:   min,
		max:   max,
	}
}

func (u *Universe) NewNumerical(rnd Rander) SingleVariate {
	return &Numerical{
		rnd: rnd,
	}
}

func (u *Universe) NewPoisson(rate float64, start, end float64) SingleVariate {
	return &Poisson{
		rate:  rate,
		value: start,
		end:   end,
		rnd:   u.Random,
	}
}

func (u *Universe) unique() string {
	u.nextvar++
	return fmt.Sprint("u", u.nextvar)
}

func (u *Universe) SomeCategories(size int) (r []ms.Category) {
	for len(r) < size {
		r = append(r, ms.Category(u.unique()))
	}
	return
}

func (c *Categorical) PDF() []float64 {
	return c.dist
}

func toDiscreteDistribution(rnd Rander, size int) DiscreteDistribution {
	cdf := make([]float64, size)
	sum := 0.0

	for i := range cdf {
		p := -1.0
		for p < 0 {
			p = rnd.Rand()
		}
		sum += p
		cdf[i] = p
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(cdf)))
	for i := 1; i < len(cdf); i++ {
		cdf[i] += cdf[i-1]
	}
	for i := range cdf {
		cdf[i] /= sum
	}

	return cdf
}

func (d *DiscreteDistribution) Pick(rnd *rand.Rand) int {
	return sort.SearchFloat64s(*d, rnd.Float64())
}

func (l *LeafProfile) Add(name ms.Variable, v SingleVariate) *LeafProfile {
	l.names = append(l.names, name)
	l.variates = append(l.variates, v)
	return l
}

func (l *LeafProfile) Produce(vs *ms.Individual) error {
	for i, v := range l.variates {
		if err := v.Generate(l.names[i], vs); err != nil {
			return err
		}
	}
	return nil
}

func (p *MixedProfile) Produce(vs *ms.Individual) error {
	number := p.dist.Pick(p.Universe.Random)
	return p.profiles[number].Produce(vs)
}

func (c *Categorical) Generate(name ms.Variable, vs *ms.Individual) error {
	number := c.dist.Pick(c.rnd)
	vs.AddCategorical(name, c.values[number])
	return nil
}

func (n *Numerical) Generate(name ms.Variable, vs *ms.Individual) error {
	number := n.rnd.Rand()
	vs.AddNumerical(name, number)
	return nil
}

func (p *Poisson) Generate(name ms.Variable, vs *ms.Individual) error {
	incr := p.rnd.ExpFloat64() / p.rate
	next := p.value + incr
	if next > p.end {
		return NotSupported
	}
	p.value = next
	vs.AddNumerical(name, p.value)
	return nil
}

func (c Categorical) Entropy() float64 {
	h := -math.Log(c.dist[0])
	for i := 1; i < len(c.dist); i++ {
		h -= math.Log(c.dist[i] - c.dist[i-1])
	}
	return h
}

func (in *InRange) Generate(name ms.Variable, vs *ms.Individual) error {
	for {
		var tmp ms.Individual
		if err := in.value.Generate(name, &tmp); err != nil {
			return err
		}

		v := tmp.Number(name)
		if v >= in.min && v <= in.max {
			vs.AddNumerical(name, v)
			return nil
		}
	}
}
