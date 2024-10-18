package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/jmacd/essay"
	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/examples/internal/multishape/continuous/tdigest"
	"github.com/jmacd/essay/examples/internal/multishape/frequency"
	"github.com/jmacd/essay/examples/internal/multishape/frequency/models"
	"github.com/jmacd/essay/examples/internal/multishape/plot"
	"github.com/jmacd/essay/examples/internal/multishape/sampler"
	"github.com/jmacd/essay/examples/internal/multishape/universe"
)

const (
	lquality tdigest.Quality = 8

	numVariables = 2.

	rate       = 10000
	duration   = 100
	colorcount = 10

	adaptivePeriods = duration
	adaptivePeriod  = duration / adaptivePeriods

	decomposePeriod = adaptivePeriod
)

var (
	greya = color.RGBA{R: 255, G: 255, B: 255, A: 128}
)

var (
	displayQuantiles = []float64{0.01, .10, .30, .50, .70, .90, .99}

	sampleSizeRatios = []float64{
		0.005,
		0.01,
		0.05,
		0.1,
	}
)

func main() {
	essay.Main("Multi sampling", multiSampling)
}

func multiSampling(doc essay.Document) {
	for _, ratio := range sampleSizeRatios {
		ratio := ratio
		doc.Section(fmt.Sprintf("2D Sample @%.2f%%", ratio*100), func(doc essay.Document) {
			twoDimensions(ratio, doc)
		})
	}
}

func twoDimensions(sampleSizeRatio float64, doc essay.Document) {
	model := models.NewModel1(plot.LatencyVar, plot.ColorVar, lquality)

	// Build synthetic data
	u := universe.New("test-2d", 499)

	randerL := u.PositiveNormal(10, 10)
	randerC := u.Normal(0, 5)

	colors := u.SomeCategories(colorcount)

	variateL := u.NewNumerical(randerL)
	variateC := u.NewCategorical(randerC, colors)

	basic := universe.NewTimeseries(u, rate, duration,
		universe.NamedVariate{plot.LatencyVar, variateL},
		universe.NamedVariate{plot.ColorVar, variateC},
	)

	var points ms.Population
	for {
		var pt ms.Individual
		if err := basic.Produce(&pt); err != nil {
			break
		}
		points = append(points, pt)
	}

	// Split the data into periods.
	var subs []ms.Population
	var sampled ms.Population

	// Split the data into `subs` subsections.
	points.Interval(universe.Time, adaptivePeriod).
		Range(0, duration).
		Aggregate(func(_, _ float64, p ms.Population) (_ ms.Individual) {
			subs = append(subs, p)
			return
		})

	scale := frequency.UniformScale()
	size := int(sampleSizeRatio*float64(len(points))/adaptivePeriods + 0.5)

	for num, subp := range subs {
		fmt.Println("-- Window", num, "size", size)

		var spoints ms.Population

		spoints, scale = sampler.Compute(subp, scale, size, model)

		sampled = append(sampled, spoints...)
	}

	summarizeWindow(fmt.Sprintf("%.2f%%", sampleSizeRatio*100), points, sampled, doc)
}

func summarizeWindow(name string, points, sampled ms.Population, doc essay.Document) {
	doc.Note("Data set", name)

	categories := points.Categorize(plot.ColorVar)
	colormap, colorCounts, colorOrder := categories.Indexed, categories.Summed, categories.Ranked

	cells := [][]interface{}{}
	top := []interface{}{"", "Original", "Sampled", "Sliced"}
	left := []interface{}{}

	// Size of data
	addLine := func(desc string, lval, rval, column interface{}) {
		left = append(left, desc)
		cells = append(cells, []interface{}{lval, rval, column})
	}

	maxLatency := points.Max(plot.LatencyVar)

	sampleCounts := sampled.Categorize(plot.ColorVar).Summed
	maxColorFrequency := math.Max(colorCounts[colorOrder[0]], sampleCounts[colorOrder[0]])

	maxRate := math.Max(
		ms.MaxY(points.Interval(universe.Time, adaptivePeriod).
			Range(0, duration).
			AsRate(plot.RateVar)),
		ms.MaxY(sampled.Interval(universe.Time, adaptivePeriod).
			Range(0, duration).
			AsRate(plot.RateVar)))

	decomposedSampled := plot.DecomposeData(sampled, categories)

	const (
		bigPlot   = 750
		smallPlot = 375
	)

	addLine("Data size", len(points), len(sampled), "")

	addLine("Latency Scatter Plot",
		plot.NewLatencyPlot(points, plot.LatencyVar).
			Interval(adaptivePeriod).
			Duration(duration).
			ColorMap(colormap).
			MaxLatency(maxLatency).
			Size(bigPlot).
			DisplayQuantiles(displayQuantiles).
			Build(),
		plot.NewLatencyPlot(sampled, plot.LatencyVar).
			Interval(adaptivePeriod).
			Duration(duration).
			ColorMap(colormap).
			MaxLatency(maxLatency).
			Size(bigPlot).
			DisplayQuantiles(displayQuantiles).
			Build(),
		plot.DecomposeGraph(decomposedSampled, func(pop ms.Population, _ int) interface{} {
			return plot.NewLatencyPlot(pop, plot.LatencyVar).
				Interval(decomposePeriod).
				Duration(duration).
				ColorMap(colormap).
				MaxLatency(maxLatency).
				Size(smallPlot).
				DisplayQuantiles(displayQuantiles).
				Build()
		}))

	addLine("Rate Scatter Plot",
		plot.NewRatePlot(points, plot.ColorVar).
			Period(adaptivePeriod).
			Duration(duration).
			ColorMap(colormap).
			MaxRate(maxRate).
			Size(bigPlot).
			Build(),
		plot.NewRatePlot(sampled, plot.ColorVar).
			Period(adaptivePeriod).
			Duration(duration).
			ColorMap(colormap).
			MaxRate(maxRate).
			Size(bigPlot).
			Build(),
		plot.DecomposeGraph(decomposedSampled, func(pop ms.Population, _ int) interface{} {
			return plot.NewRatePlot(pop, plot.ColorVar).
				Period(decomposePeriod).
				Duration(duration).
				ColorMap(colormap).
				MaxRate(maxRate).
				Size(smallPlot).
				Build()
		}))

	addLine("Color Histogram",
		plot.ColorHistogram(points, plot.ColorVar, colormap, maxColorFrequency, bigPlot, "original"),
		plot.ColorHistogram(sampled, plot.ColorVar, colormap, maxColorFrequency, bigPlot, "sampled"),
		"")

	palette := plot.HotToCoolDivergentPalette(len(colormap))

	addLine("Latency Histogram",
		plot.LatencyHistogram(points, plot.LatencyVar, maxLatency, bigPlot, greya, "original"),
		plot.LatencyHistogram(sampled, plot.LatencyVar, maxLatency, bigPlot, greya, "sampled"),
		plot.DecomposeGraph(decomposedSampled, func(pop ms.Population, num int) interface{} {

			return plot.LatencyHistogram(pop, plot.LatencyVar, 0, smallPlot, palette[num], "decomposed")
		}))

	addLine("UNWEIGHTED Color Histogram",
		plot.UnweightedColorHistogram(points, plot.ColorVar, colormap, maxColorFrequency, bigPlot, "original"),
		plot.UnweightedColorHistogram(sampled, plot.ColorVar, colormap, 0, bigPlot, "sampled"),
		"")

	addLine("UNWEIGHTED Latency Histogram",
		plot.UnweightedLatencyHistogram(points, plot.LatencyVar, palette, maxLatency, bigPlot),
		plot.UnweightedLatencyHistogram(sampled, plot.LatencyVar, palette, maxLatency, bigPlot),
		"")

	doc.Note(essay.Table{
		Cells:   cells,
		TopRow:  top,
		LeftCol: left,
	})

}
