package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/jmacd/essay"
	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/examples/internal/multishape/continuous"
	"github.com/jmacd/essay/examples/internal/multishape/continuous/tdigest"
	"github.com/jmacd/essay/examples/internal/multishape/frequency"
	"github.com/jmacd/essay/examples/internal/multishape/sampler"
	"github.com/jmacd/essay/examples/internal/multishape/universe"
	"github.com/jmacd/essay/num"
	colorful "github.com/lucasb-eyer/go-colorful"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type colorGradient [2]colorful.Color

var (
	palettes [][]colorful.Color
	black    = colorful.Color{}
	red      = colorful.Color{R: 255}
	blue     = colorful.Color{B: 255}
	green    = colorful.Color{G: 255}
	yellow   = colorful.Color{R: 255, G: 255}
	grey     = colorful.Color{R: 188, G: 188, B: 188}

	displayQuantiles = []float64{0.01, .10, .30, .50, .70, .90, .99}

	latencyGradient = colorGradient{
		red,
		green,
	}
	rateGradient = colorGradient{
		yellow,
		blue,
	}
)

const (
	maxPaletteSize = 10

	minRadius = 0.01
	lineWidth = 2

	bigPlotWidth  = 1200
	bigPlotHeight = 800

	smallPlotWidth  = 500
	smallPlotHeight = 300

	showArea = 0.05

	latencyVar ms.Variable = "latency"
	rateVar    ms.Variable = "rate"
	yposVar    ms.Variable = "ypos"

	minVarRadius = 1.75
	maxVarRadius = 2.5
)

func init() {
	palettes = make([][]colorful.Color, maxPaletteSize+1)
	for i := 1; i <= maxPaletteSize; i++ {
		palettes[i], _ = colorful.HappyPalette(i)
	}
}

func main() {
	essay.Main("Sampling Algorithms", intro)
}

func intro(doc essay.Document) {

	doc.Note(`TL;DR Honestly not for reading. Start with "Data
	Sampling", below.`)

	doc.Note(`The terms "sampling" and "sample" have many formal
uses.  We have been introduced to the idea of "drawing from" a
probability distribution, one form of _sampling_.  If a probability
distribution p(X) assigns a probability to every (continuous or
discrete) supported values of X, then drawing from the distribution
means evaluating the random variable, yielding a value, also known as
a _sample_.`,

		`_Sampling_ is also a field in statistics, where it
covers methodologies for constructing surveys and assocaited
techniques, algorithms themselves, for drawing 
inferences from a large population of individuals based a small number
of representative _samples_.  In this setting, samples are usually
multi-variate, meaning they have more than one dimension of
_explanatory_ variable.  Variables can be numerical (e.g., latency) or
categorical (e.g., service name).`,

		`Designing a sampling algorithm starts with
identifying what form of _response_ we are looking for, based on which
explanatory variables. The general approach is: considering there is a
large population of M individuals, draw a much smaller set of N
representative individuals. Based on the representatives in the
sample, we are then expected to draw inferences (i.e., estimate a
response) to the representative explanatory variables observed in the
sample data.`,

		`The term "sampling bias" describe systematic errors
in the representativity of the samples, defined in terms of response
_error_.  A "sampling estimator" is a particular sampling mechanism
with known properties in terms of bias.  An estimator could, for
example, always give the lower bound of the estimated response
variable, in which case it has a negative bias.  Sampling bias results
in inaccurate inferences.`,

		`We refer to a sampling "model" to describe the kind
of trials performed during a study, a question of intent.  If a study
aims to collect a fixed number of samples, then it is a _binomial
model_. If a study aims to collect samples over a fixed unit of time
and/or space, then it is a _Poisson model_. The mathematical
differences between these are significant, but the distinction is
largely irrelevant when sample sizes are "large enough" but still
small relative to the population size.`,

		`Sampling may be described as either _without
replacement_ or _with replacement_. referring to whether items can be
drawn at most once or possibly more than once from the population.
The mathematical differences between these are significant, but they
quantatively disappear when sample sizes are "large enough" but still
small relative to the population size.`,

		`Statistical biologists apply these tools to study
living organisms, with applications from quantifying ecosystem
diversity to evaluating medical studies.  In this setting, both
distinctions, with-replacement vs. without-replacement, and binomial
vs. poisson, takes on practical importance. Sampling estimators may be
designed for abundance sample data, in which a single study counts
total abundances (by category), vs incidence sample data, in which
repeated studies measure presence/non-presence of categories in the
sample population.`,

		`Tools have been developed for drawing correlated
inferences, using external information, for example, for correct for
incompleteness in sample data.  Sampling strategies are often tuned to
reduce bias, by applying unequal probabilities to the indivuals
selected for the sample.`)

	doc.Section("Visualizing Sample Data", computer)
}

func computer(doc essay.Document) {

	doc.Note(`Begin with a time series of two variables, frequency
	(arrival rate) and latency, an elementary stochastic process.
	There are two salient reasons why these variables are
	fundamentally different.  First, _time_ is a universal
	variable, it synchronizes the independent random processes
	(i.e., shared "current" state).  Second, these variables have
	different scales.`)

	doc.Note(`Time is has *interval* scale, meaning that
	differences are comparable.  Time-value zero is not special,
	and time can take negative values.  It is not meaningful to
	compare ratios, for example time 100 is not one tenth of time
	1000.`)

	doc.Note(`Latency has *ratio* scale, meaning we can compare
	absolute values by ratio, but that relative differences are
	relative.  It is meaningful to say latency 20 is twice as much
	as latency 10, and that the variation between 20 and 10 is
	much greater than between 120 and 110.`)

	doc.Note(`Frequency is a *count*, it has ratio scale.  Despite
	having the same kind of scale, latency and frequency are
	different kinds of variable.  We compute summary statistics of
	count variables, whereas we compute quantile statistics of
	latency variables.  Because ratio-scale differences depend on
	order-of-magnitude, it is common to plot either of them by
	logarithm.`)

	doc.Note(`The point of this is not to gain mathematical
	understanding, but to explore all the natural and
	substantially different ways of visualizing the same data,
	especially when different kinds of data are present.
	Two-dimensional plots display these three variables using X
	and Y dimensions (ignoring 3D, here), meaning that either we
	are projecting three variables onto two, or that we are using
	color (shape, texture, ...) to convey the third dimension.`)

	doc.Note(`The available visualizations are three projections
	[a] total-frequency-by-time (timeseries), [b]
	latency-quantile-by-time (timeseries), [c]
	frequency-quantile-by-latency-quantile (histogram); several
	"heatmap" colorings [d] frequency-by-time w/ latency coloring,
	[e] latency-by-time w/ frequency coloring.`)

	doc.Note(`We can also simply display the points as sample
	plots on a timeseries graph.  Presuming that the X axis
	position is time, we have two ways to project samples on the Y
	axis. The two forms are: [f] by frequency, and [g] by
	latency.`)

	doc.Note(`In the frequency sample plot [f], each sample
	represents one unit of count, and we expect to see more
	samples plotted when there is higher frequency.  Within a time
	interval, we should expect to see a number of samples plotted.
	They can be "stacked", in the sense that the samples in each
	time interval add up to the frequency during that unit of
	time.  There is no inherent connection between each sample and
	its Y position, leaving the freedom to sort samples by
	latency.`)

	doc.Note(`In the latency sample plot [g], each sample
	represents a latency value (Y), and we expect to see samples
	where they lie on the Y axis.`)

	doc.Note(`TODO: Note: the details of [f] and [g] are
	irrelevant as long as we're displaying the full data. As we
	begin to downsample to fewer points and save less data, we may
	adopt a strategy to maintain uniform coverage in [f] or [g],
	but these are different objectives.  In [f], we prefer to keep
	more samples when there is higher frequency, in [g] we prefer
	to keep more samples when there is higher spread of latencies.`)

	doc.Note(`TODO: Note: [f] corresponds to [a] in the sense that
        the stack of [f] samples rises to the [a] line on a plot with
        uniform sample density (regardless of downsampling).  [g]
        corresponds to [b] in the sense that the sample density rises
        with the latency distribution (despite downsampling).`)

	doc.Note(`Moreover, there is a between bar graphs, histograms, 
	and timeseries which gives us another way to visualize this data: 
	variable-width bar graphs.`)

	doc.Note(`TODO: Outline: 1. Show that a single "bag of spans"
	sample can generate [f] or [g] which can compute approximate
	[a], [b], [c], [d], and [e].  2. Show that we can downsample
	one continuous variable, can we estimate the errors?`)

	doc.Note(`TODO: Outline: Categorical variables. Repeat steps 1
	and 2 above.`)

	doc.Note(`TODO: Outline: Multiple variables. Repeat steps 1
	and 2 above.`)

	doc.Section(`Small`, sampleSmall)

	doc.Section(`Example`, sampleExample)

	doc.Section(`Sample Histogram`, sampleHistogram)

	doc.Section(`Sample Timeseries by Rate`, sampleTimeseriesRate)

	doc.Section(`Sample Timeseries by Latency`, sampleTimeseriesLatency)

	doc.Section(`Adaptive Timeseries Sampling`, sampleTimeseriesAdaptive)

	doc.Section(`Animated Timeseries Sampling`, sampleTimeseriesAnimated)
}

func sampleSmall(doc essay.Document) {
	const maxsize = 100000.0
	const totalSpans = 100000
	const sampsize = 10000
	names := []ms.Category{"red", "orange", "yellow", "green", "blue"}

	size := ms.Variable("size")
	samp := sampler.NewVaropt(sampsize, sampler.Intrinsic(size))
	col := ms.Variable("color")

	sum := 0
	for i := 0; i < len(names); i++ {
		sum += (i + 1)
	}
	rounds := totalSpans / sum

	actual := map[ms.Category]float64{}
	for r := 0; r < rounds; r++ {
		for i := 0; i < len(names); i++ {
			sz := maxsize / float64(i+1)
			cnt := int(maxsize / sz)

			for j := 0; j < cnt; j++ {
				var pt ms.Individual
				pt.SetAlpha(rand.Float64())
				pt.AddNumerical(size, sz)
				pt.AddCategorical(col, names[i])
				samp.Add(pt)
				actual[names[i]] += sz
			}
		}
	}
	result := map[ms.Category]float64{}
	found := map[ms.Category]float64{}
	for _, pt := range samp.Weighted() {
		cat := pt.Category(col)
		result[cat] += pt.Weight()
		found[cat]++
	}
	doc.Note(fmt.Sprintf("Tau %.4f\n", samp.Tau()))
	doc.Note(fmt.Sprintf("Total spans %d\n", samp.TotalCount()))
	doc.Note(fmt.Sprintf("Total weight %.0f\n", samp.TotalWeight()))
	doc.Note(fmt.Sprintf("Average size %.0f\n", samp.TotalWeight()/float64(samp.TotalCount())))
	for _, value := range names {
		doc.Note(fmt.Sprintf("%s has %.0f examples, total size %.0f, average size %.0f, actual total %.0f\n", value, found[value], result[value], result[value]/found[value], actual[value]))
	}
}

func sampleExample(doc essay.Document) {
	const (
		rate     = 1000.0
		duration = 10.0
	)

	u := universe.New("testing", 123)
	latencyVar := ms.Variable("latency")

	rander := u.PositiveNormal(10, 1)
	basic := universe.NewContinuousTimeseries(u, latencyVar, rander, rate, duration)
	var points ms.Population
	for {
		var pt ms.Individual
		if err := basic.Produce(&pt); err != nil {
			break
		}
		points = append(points, pt)
	}

	latencies := points.Numbers(latencyVar)

	const numBins = 3

	doc.Note(
		fmt.Sprint("Here are approximately ", rate*duration,
			" point latencies, shown in ", numBins, " variable-size bins, both axes linear."),
		num.NewPlot().
			Title("Linear X & Y").
			Size(smallPlotWidth, smallPlotHeight).
			X(num.Axis().Range(0, 20)).
			Y(num.Axis().Range(0, 0.5)).
			Legend().
			Add(num.NewHistogram().
				Normalize().
				// @@@ Note this call is having a panic, it's hard to debug!
				EqualBins(latencies, numBins).
				Name("Empirical").
				LineColor(red).
				FillColors([]color.Color{grey})).
			Add(num.Function(rander.Prob).
				Name("Theoretical").
				Samples(1000)))

	lones := points.Weights()

	digestFullRange := tdigest.Quality(numBins).New(latencies, lones,
		continuous.FullRange(),
		continuous.RangeSupported)

	digestHalfRange := tdigest.Quality(numBins).New(latencies, lones,
		continuous.PositiveRange(),
		continuous.RangeSupported)

	digestExtreme := tdigest.Quality(numBins).New(latencies, lones,
		continuous.FullRange(),
		continuous.Extremal)

	doc.Note(`Here, the purpose of computing these bins is to
		estimate the empirical probability distribution, which
		is not the same as drawing a histogram, because we
		cannot assign zero probability to any supported value.
		We have a couple of ways to extend the probability,
		either for a finite range or for an infinite range.`,

		num.NewPlot().
			Title(`Tail Behavior`).
			Size(smallPlotWidth, smallPlotHeight).
			X(num.Axis().Range(0, 20)).
			Y(num.Axis().Range(0, .5)).
			Legend().
			Add(num.Function(digestFullRange.ProbDensity).Name("Exponential").Samples(1000).Color(palettes[4][0])).
			Add(num.Function(digestHalfRange.ProbDensity).Name("Rectangular").Samples(1000).Color(palettes[4][1])).
			Add(num.Function(digestExtreme.ProbDensity).Name("Empirical").Samples(1000).Color(palettes[4][2])).
			Add(num.Function(rander.Prob).Name("Theoretical").Samples(1000).Color(palettes[4][3])))

	doc.Note(`T-digest, used to construct these bins, behaves
		asymmetrically with respect to the first and last bin.
		It tends to create a small-quantile bin on one end of
		the digest, which could be either the first or the
		last.  As written, it's the last bin which is
		generally small, so we see the more interesting
		behavior on the low end, where the first bin is a
		normal size and has to accomodate supported values
		smaller than its range.  The two behaviors differ
		depending on whether zero (polygon shape) or
		infinity (exponental shape) is the limit.`,

		num.NewPlot().
			Title(`Left Tail Behavior (Detail)`).
			Size(smallPlotWidth, smallPlotHeight).
			X(num.Axis().Range(2, 9)).
			Y(num.Axis().Range(0, .075)).
			Legend().
			Add(num.Function(digestFullRange.ProbDensity).Name("Exponential").Samples(1000).Color(palettes[4][0])).
			Add(num.Function(digestHalfRange.ProbDensity).Name("Rectangular").Samples(1000).Color(palettes[4][1])).
			Add(num.Function(digestExtreme.ProbDensity).Name("Empirical").Samples(1000).Color(palettes[4][2])).
			Add(num.Function(rander.Prob).Name("Theoretical").Samples(1000).Color(palettes[4][3])))

	samp := sampler.NewVaropt(500, frequency.NewNumberScale(digestHalfRange, latencyVar, points))

	for _, ind := range points {
		samp.Add(ind)
	}

	sampled := samp.Weighted()
	slatencies := sampled.Numbers(latencyVar)
	sweights := sampled.Weights()

	digestFromSample := tdigest.Quality(numBins).New(
		slatencies, sweights,
		continuous.PositiveRange(),
		continuous.RangeSupported)

	doc.Note(`Now, sample the distribution down to 500 points.  The histograms should look similar.`,
		num.NewPlot().
			Title(`Frequency Sample`).
			Size(smallPlotWidth, smallPlotHeight).
			X(num.Axis().Range(0, 20)).
			Y(num.Axis().Range(0, .5)).
			Legend().
			Add(num.Function(digestFromSample.ProbDensity).Name("Frequency Sampled").Samples(1000).Color(palettes[4][0])).
			Add(num.Function(digestHalfRange.ProbDensity).Name("Original").Samples(1000).Color(palettes[4][2])).
			Add(num.Function(rander.Prob).Name("Theoretical").Samples(1000).Color(palettes[4][3])))
}

func sampleHistogram(doc essay.Document) {
	u := universe.New("testing", 123)
	showSampleHistogram(doc, u, u.Exponential(1))
	showSampleHistogram(doc, u, u.PositiveNormal(1, 1))
}

func showSampleHistogram(doc essay.Document, u *universe.Universe, rander universe.Rander) {
	const (
		count = 10000
		bins  = 10
	)
	var (
		ratios = []float64{0.005, 0.01, 0.02, 0.05, 0.1}
	)

	for _, ratio := range ratios {
		latencyVar := ms.Variable("latency")

		basic := universe.NewContinuousTimeseries(u, latencyVar, rander, count, 1)
		points := universe.MakePopulation(basic)
		latencies := points.Numbers(latencyVar)
		lones := points.Weights()

		digestHalfRange := tdigest.Quality(bins).New(latencies, lones,
			continuous.PositiveRange(),
			continuous.RangeSupported)

		samp := sampler.NewVaropt(int(ratio*count), frequency.NewNumberScale(digestHalfRange, latencyVar, points))

		for _, ind := range points {
			samp.Add(ind)
		}

		sampled := samp.Weighted()
		slatencies := sampled.Numbers(latencyVar)
		sweights := sampled.Weights()

		digestFromSample := tdigest.Quality(bins).New(
			slatencies, sweights,
			continuous.PositiveRange(),
			continuous.RangeSupported)

		doc.Note(num.NewPlot().
			Title(fmt.Sprintf("%.1f%% Frequency Sample: %s", ratio*100, rander.Name())).
			Size(smallPlotWidth, smallPlotHeight).
			Range(digestHalfRange).
			Legend().
			Add(num.Function(digestFromSample.ProbDensity).Name("Frequency Sampled").Samples(1000).Color(palettes[4][0])).
			Add(num.Function(digestHalfRange.ProbDensity).Name("Original").Samples(1000).Color(palettes[4][2])).
			Add(num.Function(rander.Prob).Name("Theoretical").Samples(1000).Color(palettes[4][3])))
	}
}

func sampleTimeseriesRate(doc essay.Document) {
	u := universe.New("testing", 123)
	showSampleTimeseriesRate(doc, u, u.Exponential(1))
	showSampleTimeseriesRate(doc, u, u.PositiveNormal(1, 1))
	showSampleTimeseriesRate(doc, u, u.PositiveLogNormal(1, 1))
	showSampleTimeseriesRate(doc, u, u.Pareto(1, 1))
}

func showSampleTimeseriesRate(doc essay.Document, u *universe.Universe, rander universe.Rander) {
	const (
		duration = 10.0
		rate     = 100000.0
	)

	basic := universe.NewContinuousTimeseries(u, latencyVar, rander, rate, duration)
	points := universe.MakePopulation(basic)

	intervals := []float64{25, 50, 100, 200}

	doc.Note(fmt.Sprint("Here are ", duration, " seconds of points at a rate of ",
		rate, " per second.  Since this is a rate graph, the lines above represent the density of points filling the area beneath, with thicker lines representing average rates calculated over greater intervals.  Since there are so many points, we couldn't plot them all."),
		num.NewPlot().
			Title(fmt.Sprint(`Poisson Rate=`, rate)).
			Size(bigPlotWidth, bigPlotHeight).
			X(num.Axis().Range(0, duration)).
			Y(num.Axis()).
			Legend().
			AddEach(len(intervals), func(i int) num.Plotter {
				return num.Line(
					points.Interval(universe.Time, duration/intervals[i]).
						Range(0, duration).
						AsRate(rateVar),
				).
					Color(palettes[len(intervals)][i]).
					Name(fmt.Sprintf("interval=%.3fs", 1/intervals[i])).
					Width(200 / intervals[i])
			}))

	doc.Note(`Here, the points should fill the area, except there are too many...`,
		toRate(points, duration, 100))

	spoints := sampleLatencyOracle(points, tdigest.Quality(20), 5000)

	doc.Note(`Sampling the points down to a managable number makes them visible  points ...`,
		toRate(spoints, duration, 100))
}

type xDerivedY struct {
	points []ms.Individual
	x      ms.Variable
	y      func(i int) float64
}

func (xy xDerivedY) Len() int {
	return len(xy.points)
}

func (xy xDerivedY) XY(i int) (x, y float64) {
	return xy.points[i].Number(xy.x), xy.y(i)
}

func toRate(points ms.Population, duration float64, bins int) num.Builder {
	var ranges []ms.Population

	period := duration / float64(bins)
	points.Interval(universe.Time, period).
		Range(0, duration).
		Aggregate(func(_, _ float64, p ms.Population) (_ ms.Individual) {
			ranges = append(ranges, p)
			return
		})

	for _, bin := range ranges {
		bin.Shuffle()
		bin.SortByVar(latencyVar)
	}

	// Because of shallow copying, the addition of "ypos" here
	// isn't reflected in "scatter". Rebuild the array of combined
	// bins.
	positioned := ms.Population{}

	for _, bin := range ranges {
		sumWeight := 0.0

		for bi := range bin {
			bin[bi].AddNumerical(yposVar, sumWeight/period)
			sumWeight += bin[bi].Weight()
		}

		positioned = append(positioned, bin...)
	}

	rate := points.Interval(universe.Time, period).
		Range(0, duration).
		AsRate(rateVar)
	scatter := positioned.Points(universe.Time, yposVar)

	maxRate := ms.MaxY(rate)

	shadedArea := showArea * float64(bigPlotHeight) * float64(bigPlotWidth)
	totalWeight := points.SumWeight()
	areaWeight := shadedArea / totalWeight

	styler := func(idx int) draw.GlyphStyle {
		area := points[idx].Weight() * areaWeight
		radius := math.Sqrt(area / math.Pi)
		return draw.GlyphStyle{
			Radius: vg.Length(math.Max(radius, minRadius)),
		}
	}
	return num.NewPlot().
		Size(bigPlotWidth, bigPlotHeight).
		X(num.Axis().Range(0, duration)).
		Y(num.Axis().Range(0, maxRate)).
		Add(num.Line(rate).
			Color(blue).
			Width(lineWidth)).
		Add(num.Scatter(scatter).
			Style(styler))
}

func sampleLatencyOracle(points ms.Population, q tdigest.Quality, size int) ms.Population {
	scale, _ := sampleScale(points, q, frequency.UniformScale())
	return sampleLatency(points, size, scale)
}

func sampleScale(points ms.Population, q tdigest.Quality, prior sampler.Scale) (sampler.Scale, continuous.Digest) {
	latencies := points.Numbers(latencyVar)
	lweights := points.Weights()

	// Note: Consider building a MAP estimate here, i.e., take the
	// prior into account.

	digestHalfRange := q.New(latencies, lweights,
		continuous.PositiveRange(),
		continuous.RangeSupported)

	return frequency.NewNumberScale(digestHalfRange, latencyVar, points), digestHalfRange
}

func sampleLatency(points ms.Population, size int, scale sampler.Scale) ms.Population {
	samp := sampler.NewVaropt(size, scale)

	for _, ind := range points {
		samp.Add(ind)
	}

	return samp.Weighted()
}

func sampleTimeseriesLatency(doc essay.Document) {
	u := universe.New("testing", 123)
	showSampleTimeseriesLatency(doc, u, u.Exponential(1), false)
	showSampleTimeseriesLatency(doc, u, u.PositiveNormal(1, 1), false)
	showSampleTimeseriesLatency(doc, u, u.PositiveLogNormal(1, 1), true)
	showSampleTimeseriesLatency(doc, u, u.Pareto(1, 1), true)
}

func showSampleTimeseriesLatency(doc essay.Document, u *universe.Universe, rander universe.Rander, logscale bool) {
	const (
		duration = 10.0
		rate     = 10000.0
		periods  = 100
	)

	basic := universe.NewContinuousTimeseries(u, latencyVar, rander, rate, duration)
	points := universe.MakePopulation(basic)

	doc.Note(`This plot places points on the Y-axis at their latency value.  Here's the original data.`,
		num.NewPlot().
			Title("Latency as a scatter plot").
			Size(bigPlotWidth, bigPlotHeight).
			X(num.Axis().Range(0, duration)).
			Y(num.Axis().Min(0).LogScale(logscale)).
			Legend().
			Add(latencyQuantiles(displayQuantiles, duration/periods, points)...).
			Add(latencyFreqPoints(points)))

	spoints := sampleLatencyOracle(points, tdigest.Quality(20), 5000)

	doc.Note(`This plot places points on the Y-axis at their latency value.  Here's the sample data.`,
		num.NewPlot().
			Title("Latency as a scatter plot").
			Size(bigPlotWidth, bigPlotHeight).
			X(num.Axis().Range(0, duration)).
			Y(num.Axis().Min(0).LogScale(logscale)).
			Legend().
			Add(latencyQuantiles(displayQuantiles, duration/periods, spoints)...).
			Add(latencyFreqPoints(spoints)))
}

func latencyQuantiles(quantiles []float64, period float64, points ms.Population) (qlines []num.Plotter) {
	for i, q := range quantiles {
		qlines = append(qlines, num.Line(
			points.Interval(universe.Time, period).Aggregate(
				func(min, max float64, p ms.Population) (ind ms.Individual) {
					p.SortByVar(latencyVar)
					var item ms.Individual
					if q == 1 {
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
					ind.AddNumerical(universe.Time, (min+max)/2)
					ind.AddNumerical(latencyVar, item.Number(latencyVar))
					return
				}).Points(universe.Time, latencyVar)).
			Color(palettes[len(quantiles)][i]).
			Name(fmt.Sprintf("p%d", int(quantiles[i]*100))).
			Width(3))
	}
	return
}

func latencyFreqPoints(points ms.Population) num.Plotter {
	shadedArea := showArea * bigPlotWidth * bigPlotHeight
	totalWeight := points.SumWeight()
	areaWeight := shadedArea / totalWeight
	return num.Scatter(points.Points(universe.Time, latencyVar)).
		Style(func(pi int) draw.GlyphStyle {
			area := points[pi].Weight() * areaWeight
			radius := math.Sqrt(area / math.Pi)
			return draw.GlyphStyle{
				Radius: vg.Length(radius),
			}
		})
}

func radiusFor(min, max, value float64) vg.Length {
	if max-min == 0 {
		return vg.Points(minVarRadius)
	}
	ratio := math.Sqrt((value - min) / (max - min))
	return vg.Points(minVarRadius + ratio*(maxVarRadius-minVarRadius))
}

func sampleTimeseriesAdaptive(doc essay.Document) {
	u := universe.New("testing", 123)
	showSampleTimeseriesAdaptive(doc, u, u.Exponential(1), false)
	showSampleTimeseriesAdaptive(doc, u, u.PositiveNormal(1, 1), false)
	showSampleTimeseriesAdaptive(doc, u, u.PositiveLogNormal(1, 1), true)
	showSampleTimeseriesAdaptive(doc, u, u.Pareto(1, 1), true)
}

func showSampleTimeseriesAdaptive(doc essay.Document, u *universe.Universe, rander universe.Rander, logscale bool) {
	const (
		duration                        = 10.0
		rate                            = 100000.0
		adaptivePeriods                 = 10
		displayPeriods                  = 100
		quality         tdigest.Quality = 25
		sampleRatio                     = 0.01
		sampleSize                      = sampleRatio * (rate * duration) / adaptivePeriods
		adaptivePeriod                  = duration / adaptivePeriods
		displayPeriod                   = duration / displayPeriods
	)

	// Start with uniform probability density.
	scale := frequency.UniformScale()
	basic := universe.NewContinuousTimeseries(u, latencyVar, rander, rate, duration)
	points := universe.MakePopulation(basic)

	var subs []ms.Population
	var sampled ms.Population

	// Split the data into `subs` subsections
	points.Interval(universe.Time, adaptivePeriod).
		Range(0, duration).
		Aggregate(func(_, _ float64, p ms.Population) (_ ms.Individual) {
			subs = append(subs, p)
			return
		})

	var digests []continuous.Digest

	for _, subp := range subs {
		spoints := sampleLatency(subp, sampleSize, scale)
		sampled = append(sampled, spoints...)
		// Update the scale, passing the former scale as the prior.
		var digest continuous.Digest
		scale, digest = sampleScale(spoints, quality, scale)
		digests = append(digests, digest)
	}

	doc.Note(num.NewPlot().
		Title("Adaptive sampling: Latency").
		Size(bigPlotWidth, bigPlotHeight).
		X(num.Axis().Range(0, duration)).
		Y(num.Axis().Min(0).LogScale(logscale)).
		Legend().
		Add(latencyQuantiles(displayQuantiles, displayPeriod, sampled)...).
		Add(latencyFreqPoints(sampled)))

	doc.Note(toRate(sampled, duration, displayPeriods))

	// To see the intermediate histograms:
	// for _, d := range digests {
	// 	doc.Note(num.NewPlot().
	// 		Size(bigPlotWidth, bigPlotHeight).
	// 		X(num.Axis().Range(0, 10)).
	// 		Y(num.Axis().Range(0, 1)).
	// 		Add(num.Function(d.ProbDensity).Samples(1000)))
	// }
}

func sampleTimeseriesAnimated(doc essay.Document) {
	u := universe.New("testing", 123)
	showSampleTimeseriesAnimated(doc, u, u.Exponential(1), false)
	showSampleTimeseriesAnimated(doc, u, u.PositiveNormal(1, 1), false)
	showSampleTimeseriesAnimated(doc, u, u.PositiveLogNormal(1, 1), true)
	showSampleTimeseriesAnimated(doc, u, u.Pareto(1, 1), true)
}

func showSampleTimeseriesAnimated(doc essay.Document, u *universe.Universe, rander universe.Rander, logscale bool) {
	const (
		duration = 10.0
		rate     = 10000.0
		periods  = 100
	)

	basic := universe.NewContinuousTimeseries(u, latencyVar, rander, rate, duration)
	points := universe.MakePopulation(basic)
	var plots []essay.EncodedImage

	ratios := []float64{float64(len(points)-1) / float64(len(points)), 0.9, 0.8, 0.7, 0.6, 0.5, 0.4, .3, .2, .1, .09, .08, .07, .06, .05, .04, .03, .02, .01}

	for _, r := range ratios {
		count := int(r * float64(len(points)))

		spoints := sampleLatencyOracle(
			points,
			tdigest.Quality(25),
			count)

		plots = append(plots, num.NewPlot().
			Title(fmt.Sprintf("Sample latencies (%.0f%%)", 100*float64(count)/float64(len(points)))).
			Size(smallPlotWidth, smallPlotHeight).
			X(num.Axis().Range(0, duration)).
			Y(num.Axis().Min(0).LogScale(logscale)).
			Legend().
			Add(latencyFreqPoints(spoints)).
			Add(latencyQuantiles(displayQuantiles, duration/periods, spoints)...).
			Image(essay.PNG))
	}

	for _, p := range plots {
		doc.Note(p)
	}

	doc.Note(fmt.Sprint("Sampling down a", rander.Name()),
		essay.Animation(plots...))
}
