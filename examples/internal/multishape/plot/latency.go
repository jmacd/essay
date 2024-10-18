package plot

import (
	"fmt"
	"math"
	"sort"

	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/examples/internal/multishape/universe"
	"github.com/jmacd/essay/num"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type (
	LBuilder struct {
		v                ms.Variable
		title            string
		points           ms.Population
		duration         float64
		colormap         map[ms.Category]int
		maxLatency       float64
		showPoints       bool
		size             int
		interval         float64
		displayQuantiles []float64
	}
)

func NewLatencyPlot(points ms.Population, v ms.Variable) LBuilder {
	return LBuilder{
		points:   points,
		v:        v,
		interval: 1,
	}
}

func (l LBuilder) Duration(duration float64) LBuilder {
	l.duration = duration
	return l
}

func (l LBuilder) Title(title string) LBuilder {
	l.title = title
	return l
}

func (l LBuilder) MaxLatency(maxLatency float64) LBuilder {
	l.maxLatency = maxLatency
	return l
}

func (l LBuilder) ShowPoints(showP bool) LBuilder {
	l.showPoints = showP
	return l
}

func (l LBuilder) Size(size int) LBuilder {
	l.size = size
	return l
}

func (l LBuilder) DisplayQuantiles(displayQuantiles []float64) LBuilder {
	l.displayQuantiles = displayQuantiles
	return l
}

func (l LBuilder) ColorMap(colormap map[ms.Category]int) LBuilder {
	l.colormap = colormap
	return l
}

func (l LBuilder) Interval(i float64) LBuilder {
	l.interval = i
	return l
}

func (l LBuilder) latencyQuantiles() (qlines []num.Plotter) {
	for i, q := range l.displayQuantiles {
		points := l.points.Interval(universe.Time, l.interval).
			Range(-l.duration/4, l.duration*5/4).
			Aggregate(
				func(min, max float64, p ms.Population) (ind ms.Individual) {
					ind.AddNumerical(universe.Time, (min+max)/2)
					if len(p) == 0 {
						return
					}
					p.SortByVar(l.v)
					var item ms.Individual
					if q == 0 {
						item = p[0]
					} else if q == 1 {
						item = p[len(p)-1]
					} else {
						wsum, wcum := p.CumWeight()
						wval := q * wsum
						pos := sort.Search(len(p), func(i int) bool {
							return wcum[i] >= wval
						})
						if pos >= len(wcum) || (wcum[pos] != wval && pos != 0) {
							pos--
						}
						item = p[pos]
					}
					ind.AddNumerical(l.v, item.Number(l.v))
					return
				}).Points(universe.Time, l.v)

		spoints := ValueSmoothing(points)

		line := num.Line(spoints).
			Smooth().
			Name(fmt.Sprintf("p%d", int(l.displayQuantiles[i]*100))).
			Width(lineWidth)

		qlines = append(qlines, line)
	}
	return
}

func (l LBuilder) latencyFreqPoints() num.Plotter {
	shadedArea := showArea * float64(l.size) * float64(l.size)
	totalWeight := l.points.SumWeight()
	areaWeight := shadedArea / totalWeight
	return num.Scatter(l.points.Points(universe.Time, l.v)).
		Style(func(pi int) draw.GlyphStyle {
			area := l.points[pi].Weight() * areaWeight
			radius := math.Sqrt(area / math.Pi)
			gs := draw.GlyphStyle{
				Radius: vg.Length(math.Max(radius, minRadius)),
			}
			if l.colormap != nil {
				gs.Color = HotToCoolDivergentPalette(len(l.colormap))[l.colormap[l.points[pi].Category(ColorVar)]]
			}
			return gs
		})
}

func (l LBuilder) Build() num.Builder {
	if l.maxLatency == 0 {
		l.maxLatency = l.points.Max(l.v)
	}
	r := num.NewPlot().
		Size(l.size, l.size).
		Title(l.title).
		X(num.Axis().Range(0, l.duration).Hide()).
		Y(num.Axis().Range(0, l.maxLatency).LogScale(false).Hide())
	if l.showPoints {
		r = r.Add(l.latencyFreqPoints())
	}
	r = r.Add(l.latencyQuantiles()...)
	return r
}
