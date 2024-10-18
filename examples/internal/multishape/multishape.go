package multishape

import (
	"fmt"
	"math"
	"sort"
)

// const (
// 	CountMeasure    MeasureType = iota // Counts are logarithmically scaled
// 	IntervalMeasure                    // Equal spacing w/o true zero, e.g., time
// 	RatioMeasure                       // Proportional spacing w/ true zero, e.g., speed, latency
// )

type (
	Variable string
	Category string

	// Note: This data set is inefficient to access, but cheap to
	// build. Some kind of report signature will be needed to make
	// this efficient.
	VariableSet struct {
		numv []Variable
		nums []float64

		catv []Variable
		cats []Category
	}

	Individual struct {
		// Input
		alpha float64
		vars  VariableSet

		// Computed
		priority float64
		weight   float64
	}

	Population []Individual

	// XYer wraps the Len and XY methods (for gonum/plot).
	XYer interface {
		// Len returns the number of x, y pairs.
		Len() int

		// XY returns an x, y pair.
		XY(int) (x, y float64)

		XYok(int) (x, y float64, ok bool)
	}
)

func (ind *Individual) AddCategorical(name Variable, cat Category) {
	ind.vars.catv = append(ind.vars.catv, name)
	ind.vars.cats = append(ind.vars.cats, cat)
}

func (ind *Individual) AddNumerical(name Variable, num float64) {
	ind.vars.numv = append(ind.vars.numv, name)
	ind.vars.nums = append(ind.vars.nums, num)
}

func (ind *Individual) Number(name Variable) float64 {
	if v, ok := ind.GetNumber(name); ok {
		return v
	}
	panic(fmt.Sprint("Missing variable: ", name, ": ", *ind))
}

func (ind *Individual) GetNumber(name Variable) (float64, bool) {
	for i, v := range ind.vars.numv {
		if v == name {
			return ind.vars.nums[i], true
		}
	}
	return 0, false
}

func (ind *Individual) Category(name Variable) Category {
	for i, v := range ind.vars.catv {
		if v == name {
			return ind.vars.cats[i]
		}
	}
	panic(fmt.Sprint("Missing variable: ", name, ": ", *ind))
}

func (ind *Individual) SetAlpha(alpha float64) {
	ind.alpha = alpha
	ind.setPriority()
}

func checkf(p *Individual, f float64) {
	if f < 0 {
		panic(fmt.Sprint("Impossible: ", *p, ": ", f))
	}
	if f > 100e100 {
		panic(fmt.Sprint("Insane: ", *p, ": ", f))
	}
	if math.IsNaN(f) {
		panic(fmt.Sprint("NaN: ", *p, ": ", f))
	}
	if math.IsInf(f, 0) {
		panic(fmt.Sprint("Inf: ", *p, ": ", f))
	}
}

func (ind *Individual) SetWeight(weight float64) {
	checkf(ind, weight)
	ind.weight = weight
	ind.setPriority()
}

func (ind *Individual) setPriority() {
	if ind.alpha != 0.0 {
		ind.priority = ind.weight / ind.alpha
	} else {
		ind.priority = math.NaN()
	}
}

func (ind *Individual) Priority() float64 {
	return ind.priority
}

func (ind *Individual) Weight() float64 {
	return ind.weight
}

type Categories struct {
	Var     Variable
	Indexed map[Category]int
	Summed  map[Category]float64
	Ranked  []Category
}

// func (c Categories) GetIndex(cat Category) int {
// 	return c.Indexed[cat]
// }

// func (c Categories) Get(cat Category) int {
// 	return c.Indexed[cat]
// }

// Assigns categories their rank in descending order by frequency.
// Returns a map from color name to color number, a map from color
// name to total count, and as a slice by rank.
func (points Population) Categorize(catVar Variable) Categories {
	catSlice, catCounts := points.CategoryCounts(catVar)

	sort.Slice(catSlice, func(i, j int) bool {
		return catCounts[catSlice[i]] > catCounts[catSlice[j]]
	})

	catmap := map[Category]int{}
	for i, c := range catSlice {
		catmap[c] = i
	}
	return Categories{
		Var:     catVar,
		Indexed: catmap,
		Summed:  catCounts,
		Ranked:  catSlice,
	}
}

func CategoryMap(values []Category) map[Category]int {
	r := map[Category]int{}
	for i, v := range values {
		r[v] = i
	}
	return r
}

func (points Population) CategoryCounts(catVar Variable) ([]Category, map[Category]float64) {
	catCounts := map[Category]float64{}
	for _, p := range points {
		catCounts[p.Category(catVar)] += p.Weight()
	}

	catSlice := make([]Category, 0, len(catCounts))
	for c, _ := range catCounts {
		catSlice = append(catSlice, c)
	}
	return catSlice, catCounts
}
