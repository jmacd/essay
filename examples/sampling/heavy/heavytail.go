package main

import (
	"fmt"
	"math"

	"github.com/jmacd/essay"
	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/examples/internal/multishape/frequency"
	"github.com/jmacd/essay/examples/internal/multishape/frequency/models"
	"github.com/jmacd/essay/examples/internal/multishape/plot"
	"github.com/jmacd/essay/examples/internal/multishape/sampler"
	"github.com/jmacd/essay/examples/internal/multishape/universe"
)

const (
	rate     = 10000
	duration = 100

	adaptivePeriods = duration
	adaptivePeriod  = duration / adaptivePeriods

	decomposePeriod = adaptivePeriod
)

var (
	// sample size is relative to (rate * duration / adaptivePeriods).
	sampleSizeRatios = []float64{
		0.01, // 100
		0.05, // 500
		0.1,  // 1000
	}

	// color count is relative to sample size.
	colorCountRatios = []float64{
		0.25,
		0.5,
		1,
		2,
		5,
		10,
	}
)

func main() {
	essay.Main("Heavy tail", heavytailSampling)
}

func heavytailSampling(doc essay.Document) {
	for _, cratio := range colorCountRatios {
		cratio := cratio
		for _, sratio := range sampleSizeRatios {
			sratio := sratio

			doc.Section(
				fmt.Sprintf("Heavy 1D Colors @%.2f%% Sample @%.2f%%", cratio*100, sratio*100),
				func(doc essay.Document) {
					heavySample(cratio, sratio, doc)
				})
		}
	}
}

func heavySample(cratio, sratio float64, doc essay.Document) {
	model := models.NewModel2(plot.ColorVar)

	// Build synthetic data
	u := universe.New("heavy-1d", 1723)

	pointsPerPeriod := float64(duration*rate) / adaptivePeriods
	sampleSize := int(sratio * pointsPerPeriod)
	colorCount := int(cratio * sratio * pointsPerPeriod)

	randerC := u.Exponential(5)
	colorList := u.SomeCategories(colorCount)
	variateC := u.NewCategorical(randerC, colorList)
	colorOrder := colorList
	colorMap := ms.CategoryMap(colorOrder)

	fmt.Println("Expt with points/period", pointsPerPeriod, "with", len(colorOrder), "colors")

	basic := universe.NewTimeseries(u, rate, duration,
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

	distributionEntropy := variateC.(*universe.Categorical).Entropy()
	populationEntroppy := computeEntropy(points)
	populationDistinct := countDistinct(points)

	var subpopEntropy []float64
	var sampleEntropy []float64
	var subpopDistinct []int
	var sampleDistinct []int

	for _, subp := range subs {
		var spoints ms.Population

		spoints, scale = sampler.Compute(subp, scale, sampleSize, model)

		subpopEntropy = append(subpopEntropy, computeEntropy(subp))
		sampleEntropy = append(sampleEntropy, computeEntropy(spoints))
		subpopDistinct = append(subpopDistinct, countDistinct(subp))
		sampleDistinct = append(sampleDistinct, countDistinct(spoints))

		sampled = append(sampled, spoints...)
	}

	summarizeWindow(fmt.Sprintf("sample %.2f%% colors %.2f%%", sratio*100, cratio*100),
		points, sampled, colorMap, colorOrder,
		distributionEntropy, populationEntroppy, subpopEntropy, sampleEntropy,
		colorCount, populationDistinct, subpopDistinct, sampleDistinct,
		doc)
}

func countDistinct(points ms.Population) int {
	m := map[ms.Category]struct{}{}
	for _, p := range points {
		m[p.Category(plot.ColorVar)] = struct{}{}
	}
	return len(m)
}

func computeEntropy(points ms.Population) float64 {
	sum := 0.0
	for _, p := range points {
		sum += p.Weight()
	}
	h := 0.0
	for _, p := range points {
		h -= math.Log(p.Weight() / sum)
	}
	return h
}

func summarizeWindow(
	name string,
	points, sampled ms.Population,
	colorMap map[ms.Category]int,
	colorOrder []ms.Category,
	distributionEntropy, populationEntropy float64,
	subpopEntropy, sampleEntropy []float64,
	distributionDistinct, populationDistinct int,
	subpopDistinct, sampleDistinct []int,
	doc essay.Document,
) {
	doc.Note("Data set", name)

	_, colorCounts := points.CategoryCounts(plot.ColorVar)

	cells := [][]interface{}{}
	top := []interface{}{"", "Original", "Sampled", "Sliced"}
	left := []interface{}{}

	addLine := func(desc string, lval, rval, column interface{}) {
		left = append(left, desc)
		cells = append(cells, []interface{}{lval, rval, column})
	}

	_, sampleCounts := sampled.CategoryCounts(plot.ColorVar)

	maxColorFrequency := math.Max(colorCounts[colorOrder[0]], sampleCounts[colorOrder[0]])

	maxRate := math.Max(
		ms.MaxY(points.Interval(universe.Time, adaptivePeriod).
			Range(0, duration).
			AsRate(plot.RateVar)),
		ms.MaxY(sampled.Interval(universe.Time, adaptivePeriod).
			Range(0, duration).
			AsRate(plot.RateVar)))

	const (
		bigPlot   = 750
		smallPlot = 375
	)

	addLine("Data size", len(points), len(sampled), "")

	addLine("Rate Scatter Plot",
		// This call is expensive, has 1M points to plot
		// plot.NewRatePlot(points, plot.ColorVar).
		// 	Period(adaptivePeriod).
		// 	Duration(duration).
		// 	ColorMap(colorMap).
		// 	MaxRate(maxRate).
		// 	Size(bigPlot).
		// 	Build(),
		"",
		plot.NewRatePlot(sampled, plot.ColorVar).
			Period(adaptivePeriod).
			Duration(duration).
			ColorMap(colorMap).
			MaxRate(maxRate).
			Size(bigPlot).
			Build(),
		"",
	)

	addLine("Color Histogram",
		plot.ColorHistogram(points, plot.ColorVar, colorMap, maxColorFrequency, bigPlot, "original"),
		plot.ColorHistogram(sampled, plot.ColorVar, colorMap, maxColorFrequency, bigPlot, "sampled"),
		"")

	addLine("UNWEIGHTED Color Histogram",
		plot.UnweightedColorHistogram(points, plot.ColorVar, colorMap, maxColorFrequency, bigPlot, "origina"),
		plot.UnweightedColorHistogram(sampled, plot.ColorVar, colorMap, 0, bigPlot, "sampled"),
		"")

	doc.Note(essay.Table{
		Cells:   cells,
		TopRow:  top,
		LeftCol: left,
	})

}

// num.NewPlot().
// 	Title(fmt.Sprint(`Poisson Rate=`, rate)).
// 	Size(bigPlotWidth, bigPlotHeight).
// 	X(num.Axis().Range(0, duration)).
// 	Y(num.Axis()).
// 	Legend().
// 	AddEach(len(intervals), func(i int) num.Plotter {
// 		return num.Line(
// 			points.Interval(universe.Time, duration/intervals[i]).
// 				Range(0, duration).
// 				AsRate(rateVar),
// 		).
// 			Color(palettes[len(intervals)][i]).
// 			Name(fmt.Sprintf("interval=%.3fs", 1/intervals[i])).
// 			Width(200 / intervals[i])
// 	}))
