package plot

import (
	"math"

	"github.com/jmacd/essay"
	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/num"
)

const (
	LatencyVar ms.Variable = "latency"
	ColorVar   ms.Variable = "color"
	RateVar    ms.Variable = "rate"
	YposVar    ms.Variable = "ypos"
)

func DecomposeData(points ms.Population, categories ms.Categories) []ms.Population {
	numCats := len(categories.Indexed)

	decomposed := make([]ms.Population, numCats)

	for _, pt := range points {
		catCol := categories.Indexed[pt.Category(categories.Var)]
		decomposed[catCol] = append(decomposed[catCol], pt)
	}

	return decomposed
}

func DecomposeGraph(pops []ms.Population, f func(ms.Population, int) interface{}) interface{} {
	cells := make([][]interface{}, 2)
	for i := range cells {
		cells[i] = make([]interface{}, len(pops)/2)
	}
	for i, pop := range pops {
		cells[i%2][i/2] = f(pop, i)
	}
	return essay.Table{
		Cells: cells,
	}
}

func RateSmoothing(
	points ms.Population,
	v, dv ms.Variable,
	duration,
	interval float64,
) num.XYs {
	const minPoints = 3

	raw := points.Interval(v, interval).
		Range(-duration, 2*duration).
		AsRate(dv)

	return smooth(raw)
}

func ValueSmoothing(raw ms.XYer) num.XYs {
	return smooth(raw)
}

func smooth(raw ms.XYer) num.XYs {
	const minPoints = 3

	var res num.XYs
	rough := math.Inf(+1)
	kurt := 0.0

	for ws := 1; ws < raw.Len()/minPoints; ws++ {
		sma := num.XYs{}

		for i := 0; i <= raw.Len()-ws; i++ {
			xt := 0.0
			yt := 0.0
			for j := 0; j < ws; j++ {
				x, y, _ := raw.XYok(i + j)
				xt += x
				yt += y
			}
			xr := xt / float64(ws)
			yr := yt / float64(ws)
			sma.X = append(sma.X, xr)
			sma.Y = append(sma.Y, yr)
		}

		r := roughness(sma)
		k := kurtosis(sma)

		if r < rough && k > kurt {
			rough = r
			kurt = k
			res = sma
		}
	}

	return res
}

func roughness(d num.XYs) float64 {
	r := 0.0
	for i := 1; i < d.Len(); i++ {
		_, y1 := d.XY(i - 1)
		_, y2 := d.XY(i)

		r += math.Abs(y2 - y1)
	}

	return r / float64(d.Len()-1)
}

func kurtosis(d num.XYs) float64 {
	s := 0.0
	for i := 0; i < d.Len(); i++ {
		_, y := d.XY(i)
		s += y
	}
	mean := s / float64(d.Len())

	knum := 0.0
	kden := 0.0
	for i := 0; i < d.Len(); i++ {
		_, y := d.XY(i)
		knum += math.Pow(y-mean, 4)
		kden += math.Pow(y-mean, 2)
	}
	knum /= float64(d.Len())
	kden /= float64(d.Len())

	return knum / (kden * kden)
}
