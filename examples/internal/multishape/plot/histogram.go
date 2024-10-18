package plot

import (
	"fmt"
	"image/color"
	"math"

	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/num"
	"gonum.org/v1/plot/plotter"
)

const (
	lineWidth = 3

	minRadius = 0.01

	// TODO variable
	showHistoBins = 30

	showArea = 0.25
)

func ColorHistogram(points ms.Population, v ms.Variable, colormap map[ms.Category]int, maxColorFrequency float64, sz int, title string) num.Builder {
	return ColorHistogramBy(points, v, colormap, maxColorFrequency, sz,
		func(p ms.Individual) float64 { return p.Weight() }, title)
}

func UnweightedColorHistogram(points ms.Population, v ms.Variable, colormap map[ms.Category]int, maxColorFrequency float64, sz int, title string) num.Builder {
	return ColorHistogramBy(points, v, colormap, maxColorFrequency, sz,
		func(p ms.Individual) float64 { return 1 }, title)
}

func ColorHistogramBy(points ms.Population, v ms.Variable, colormap map[ms.Category]int, maxColorFrequency float64, sz int, by func(p ms.Individual) float64, title string) num.Builder {
	counts := make(plotter.Values, len(colormap))
	for _, p := range points {
		counts[colormap[p.Category(v)]] += by(p)
	}
	names := make([]string, len(colormap))
	for i := range names { // TODO: name these?
		names[i] = fmt.Sprint("i", i)
	}

	for _, c := range counts {
		maxColorFrequency = math.Max(c, maxColorFrequency)
	}

	pal := HotToCoolDivergentPalette(len(colormap))

	return num.NewPlot().
		Size(sz, sz).
		Title(title).
		X(num.Axis().Nominal(names).Hide()).
		Y(num.Axis().Range(0, maxColorFrequency).LogScale(false).Hide()).
		Add(num.NominalHistogram(counts, func(i int) color.Color {
			return pal[i]
		}))
}

// func LatencyHistogram(points ms.Population, v ms.Variable, maxLatency float64, sz int) num.Builder {
// 	return num.NewPlot().
// 		Size(sz, sz).
// 		X(num.Axis().Min(0).Max(maxLatency)).
// 		Y(num.Axis().Min(0)).
// 		Add(num.WeightedHistogram(points.WeightedNumbers(v)).
// 			Bins(showHistoBins).
// 			LineColor(Black).
// 			FillColor(CoolToHotSequentialPalette(showHistoBins)))
// }

// func UnweightedLatencyHistogram(points ms.Population, v ms.Variable, colormap []color.Color, maxLatency float64, sz int) num.Builder {
// 	return num.NewPlot().
// 		Size(sz, sz).
// 		X(num.Axis().Min(0).Max(maxLatency)).
// 		Y(num.Axis().Min(0)).
// 		Add(num.Float64Histogram(points.Numbers(v)).
// 			Bins(showHistoBins).
// 			LineColor(Black).
// 			FillColor(CoolToHotSequentialPalette(showHistoBins)))
// }
