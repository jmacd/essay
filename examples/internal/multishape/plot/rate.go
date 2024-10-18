package plot

import (
	"math"

	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/examples/internal/multishape/universe"
	"github.com/jmacd/essay/num"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type (
	RBuilder struct {
		points   ms.Population
		v        ms.Variable
		duration float64
		period   float64
		maxRate  float64
		colormap map[ms.Category]int
		size     int
	}
)

func NewRatePlot(points ms.Population, v ms.Variable) RBuilder {
	return RBuilder{
		points: points,
		v:      v,
	}
}

func (r RBuilder) Period(period float64) RBuilder {
	r.period = period
	return r
}

func (r RBuilder) Duration(duration float64) RBuilder {
	r.duration = duration
	return r
}

func (r RBuilder) MaxRate(maxRate float64) RBuilder {
	r.maxRate = maxRate
	return r
}

func (r RBuilder) Size(size int) RBuilder {
	r.size = size
	return r
}

func (r RBuilder) ColorMap(colormap map[ms.Category]int) RBuilder {
	r.colormap = colormap
	return r
}

func (r RBuilder) Build() num.Builder {
	var ranges []ms.Population

	r.points.Interval(universe.Time, r.period).
		Range(0, r.duration).
		Aggregate(func(_, _ float64, p ms.Population) (_ ms.Individual) {
			ranges = append(ranges, p)
			return
		})

	for _, bin := range ranges {
		bin.Shuffle()

		bin.SortBy(func(p ms.Individual) float64 {
			return float64(r.colormap[p.Category(r.v)])
		})
	}

	// Because of shallow copying, the addition of "ypos" here
	// isn't reflected in "scatter". Rebuild the array of combined
	// bins.
	positioned := ms.Population{}

	for _, bin := range ranges {
		sumWeight := 0.0

		for bi := range bin {
			bin[bi].AddNumerical(YposVar, sumWeight)
			sumWeight += bin[bi].Weight()
		}

		positioned = append(positioned, bin...)
	}

	rate := r.points.Interval(universe.Time, r.period).
		Range(0, r.duration).
		AsRate(RateVar)
	scatter := positioned.Points(universe.Time, YposVar)

	if r.maxRate == 0 {
		r.maxRate = ms.MaxY(rate)
	}

	shadedArea := showArea * float64(r.size) * float64(r.size)
	totalWeight := r.points.SumWeight()
	areaWeight := shadedArea / totalWeight

	styler := func(idx int) draw.GlyphStyle {
		area := r.points[idx].Weight() * areaWeight
		radius := math.Sqrt(area / math.Pi)
		return draw.GlyphStyle{
			Radius: vg.Length(math.Max(radius, minRadius)),
			Color:  HotToCoolDivergentPalette(len(r.colormap))[r.colormap[positioned[idx].Category(r.v)]],
		}
	}

	return num.NewPlot().
		Size(r.size, r.size).
		X(num.Axis().Range(0, r.duration)).
		Y(num.Axis().Range(0, r.maxRate)).
		Add(num.Line(rate).
			Color(Black).
			Width(lineWidth)).
		Add(num.Scatter(scatter).
			Style(styler))
}
