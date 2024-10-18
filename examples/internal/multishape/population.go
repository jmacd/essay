package multishape

import (
	"log"
	"math"
	"runtime/debug"
	"sort"

	"golang.org/x/exp/rand"
)

type (
	IBuilder struct {
		population Population
		min        float64
		max        float64
		input      Variable
		interval   float64
	}

	Aggregator func(min, max float64, p Population) Individual

	pointXY struct {
		population Population
		x, y       func(Individual) (float64, bool)
	}

	populationSorter struct {
		p  Population
		vf func(Individual) float64
	}
)

func (p Population) Interval(input Variable, interval float64) IBuilder {
	return IBuilder{
		population: p.Copy(),
		min:        math.Inf(+1),
		max:        math.Inf(-1),
		interval:   interval,
		input:      input,
	}
}

func (i IBuilder) Range(min, max float64) IBuilder {
	i.min = min
	i.max = max
	return i
}

func (i IBuilder) AsCount(output Variable) XYer {
	return i.selection(i.countAs(output, 1)).Points(i.input, output)
}

func (i IBuilder) AsRate(output Variable) XYer {
	return i.selection(i.countAs(output, 1/i.interval)).Points(i.input, output)
}

func (i IBuilder) Aggregate(agg Aggregator) Population {
	return i.selection(agg)
}

func (i IBuilder) countAs(output Variable, multiplier float64) Aggregator {
	return func(min, max float64, p Population) (ind Individual) {
		ind.AddNumerical(i.input, (min+max)/2)
		ind.AddNumerical(output, p.SumWeight()*multiplier)
		return
	}
}

func MaxY(data XYer) float64 {
	r := math.Inf(-1)
	for i := 0; i < data.Len(); i++ {
		_, y := data.XY(i)
		r = math.Max(r, y)
	}
	return r
}

func (i IBuilder) selection(agg Aggregator) (result Population) {
	i.population.SortByVar(i.input)
	size := len(i.population)
	if size > 0 {
		if math.IsInf(i.min, +1) {
			i.min = i.population[0].Number(i.input)
		}
		if math.IsInf(i.max, -1) {
			i.max = i.population[size-1].Number(i.input)
		}
	}

	if i.interval == 0 {
		panic("Invalid interval")
	}

	position := 0
	boundary := i.min + i.interval
	for boundary <= i.max {
		var sub Population

		for position < size {
			pt := i.population[position]
			val := pt.Number(i.input)

			if val < boundary {
				sub = append(sub, pt)
				position++
				continue
			}

			break
		}

		out := agg(boundary-i.interval, boundary, sub)
		result = append(result, out)
		boundary += i.interval
	}
	return
}

func (c pointXY) XY(i int) (x, y float64) {
	pt := c.population[i]
	x, xok := c.x(pt)
	y, yok := c.y(pt)
	if !xok || !yok {
		log.Fatal("RAR", string(debug.Stack()))
		panic("Invalid XY")
	}
	return x, y
}
func (c pointXY) XYok(i int) (x, y float64, ok bool) {
	pt := c.population[i]
	x, xok := c.x(pt)
	y, yok := c.y(pt)
	return x, y, xok && yok
}
func (c pointXY) Len() int {
	return len(c.population)
}

func (p Population) SortBy(vf func(Individual) float64) {
	sort.Sort(populationSorter{
		p:  p,
		vf: vf,
	})
}

func (p Population) SortByVar(v Variable) {
	p.SortBy(func(p Individual) float64 {
		return p.Number(v)
	})
}

func (ps populationSorter) Len() int {
	return len(ps.p)
}
func (ps populationSorter) Swap(i, j int) {
	ps.p[i], ps.p[j] = ps.p[j], ps.p[i]
}
func (ps populationSorter) Less(i, j int) bool {
	return ps.vf(ps.p[i]) < ps.vf(ps.p[j])
}

func (p Population) Points(x, y Variable) XYer {
	return pointXY{
		population: p,
		x:          func(p Individual) (float64, bool) { return p.GetNumber(x) },
		y:          func(p Individual) (float64, bool) { return p.GetNumber(y) },
	}
}

func (p Population) Weights() (w []float64) {
	w = make([]float64, 0, len(p))
	for _, pt := range p {
		w = append(w, pt.Weight())
	}
	return
}

func (p Population) SumWeight() (w float64) {
	for _, pt := range p {
		w += pt.Weight()
	}
	return
}

func (p Population) Numbers(v Variable) (r []float64) {
	r = make([]float64, len(p))
	for i := range p {
		r[i] = p[i].Number(v)
	}
	return
}

func (p Population) WeightedNumbers(v Variable) XYer {
	return pointXY{
		population: p,
		x:          func(p Individual) (float64, bool) { return p.GetNumber(v) },
		y:          func(p Individual) (float64, bool) { return p.Weight(), true },
	}
}

func (p Population) Max(v Variable) float64 {
	r := math.Inf(-1)
	for _, pt := range p {
		r = math.Max(r, pt.Number(v))
	}
	return r
}

func (p Population) Min(v Variable) float64 {
	r := math.Inf(+1)
	for _, pt := range p {
		r = math.Min(r, pt.Number(v))
	}
	return r
}

func (p Population) Categories(v Variable) (r []Category) {
	r = make([]Category, len(p))
	for i := range p {
		r[i] = p[i].Category(v)
	}
	return
}

func (p Population) Copy() Population {
	// NOTE: Shallow copy.
	c := make(Population, len(p))
	copy(c, p)
	return c
}

func (p Population) Filter(filter func(Individual) bool) Population {
	var c Population

	for _, pt := range p {
		if filter(pt) {
			c = append(c, pt)
		}
	}

	return c
}

func (p Population) Shuffle() {
	rand.Shuffle(len(p), func(i, j int) {
		p[i], p[j] = p[j], p[i]
	})
}

func (p Population) CumWeight() (sum float64, cum []float64) {
	for _, pt := range p {
		sum += pt.Weight()
		cum = append(cum, sum)
	}
	return
}

func (p Population) MaxWeight() (w float64) {
	w = math.Inf(-1)
	for _, pt := range p {
		w = math.Max(w, pt.Weight())
	}
	return
}

func (p Population) MinWeight() (w float64) {
	w = math.Inf(+1)
	for _, pt := range p {
		w = math.Min(w, pt.Weight())
	}
	return
}

func (p Population) MaxNumber(v Variable) (w float64) {
	w = math.Inf(-1)
	for _, pt := range p {
		w = math.Max(w, pt.Number(v))
	}
	return
}

func (p Population) MinNumber(v Variable) (w float64) {
	w = math.Inf(+1)
	for _, pt := range p {
		w = math.Min(w, pt.Number(v))
	}
	return
}
