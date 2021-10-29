package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/lightstep/sandbox/jmacd/essay"
	"github.com/lightstep/sandbox/jmacd/essay/num"
	"github.com/lightstep/sandbox/jmacd/gonum/loghist"
	"gonum.org/v1/gonum/stat/distuv"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var (
	blue  = color.RGBA{B: 255, A: 255}
	red   = color.RGBA{R: 255, A: 255}
	green = color.RGBA{G: 255, A: 255}
	greya = color.RGBA{R: 255, G: 255, B: 255, A: 128}
	greys = color.RGBA{R: 50, G: 80, B: 70, A: 255}
	black = color.RGBA{R: 0, G: 0, B: 0, A: 255}
)

const (
	X axis = iota
	Y
)

type (
	axis int

	DistAPI interface {
		Rand() float64
		Mean() float64
		Mode() float64
		Prob(float64) float64
		CDF(float64) float64
	}

	GoodDistAPI interface {
		DistAPI
		Quantile(float64) float64
	}

	Dist interface {
		GoodDistAPI
		String() string
	}

	HeavyDist interface {
		DistAPI
		String() string
	}

	gonumDist struct {
		GoodDistAPI
		name string
	}
	heavyDist struct {
		DistAPI
		name string
	}

	LineGlyph struct{}

	XY1D []float64
)

func main() {
	essay.Main("Univariate Distributions", write)
}

func write(doc essay.Document) {
	doc.Note(`This essay uses gonum/plot and gonum/stat/distuv to
		display CDFs, PDFs, histograms, and samples drawn from
		univariate distributions.`)

	doc.Section("Probability Distributions", noteDistributions)

	doc.Section("Displaying Random Varibles", noteDisplaying)

	doc.Section("Histograms of Continuous Variables", noteHistograms)

	doc.Section("Skewed Distributions", noteSkewed)

	doc.Section("Heavy-Tail Distributions", noteHeavyTail)
}

func noteDistributions(doc essay.Document) {

	doc.Note(`A random variable is a variable whose values are the
		outcome of a random phenomenon. Random variables are described 
		by their probability distribution.`,

		`A continuous probability distribution function maps
		supported inputs to probability values in the range [0,1) such
		that the total probability sum is 1.`)

	doc.Note(`The standard normal distribution is centered at 0 and
		has unit variance.  This shows 99% of the total 
		probability distribution function (PDF).  The mathematical
		expression p(x) refers to the probability of drawing x from
		a random variable with a certain distribution.`,
		plot1(normalDist(0, 1), red, 0.99))

	doc.Note(`We can change the mean and the variance without
		changing the shape of the distribution.  The mean and
		variance are called "location" and "scale"
		parameters.`, plot1(normalDist(10, 10), blue, 0.99))

	doc.Note(`Place them on the same plot to see the difference in 
		location and scale.`,
		plot2("Bi-modal distribution",
			normalDist(0, 1),
			normalDist(10, 10),
			red, blue,
			0.99))

	doc.Note(`The normal distribution extends to infinity in both 
		directions, whereas the Gamma distribution supports only
		positive values.`,
		plot1(gammaDist(2, 2), green, 0.99))

	doc.Note(`The cumulative distribution function describes the 
		 sum of probability that is less-than a certain supported
		 value. It can also be called an inverse-quantile function, 
		 because it maps values into quantiles of the distribution.`,
		plotCDF(normalDist(0, 1), red, 0.99))

	doc.Note(`This is the scaled and shifted normal distribution's CDF.`,
		plotCDF(normalDist(10, 10), blue, 0.99))

	doc.Note(`The CDF for the Gamma distribution shown above.`,
		plotCDF(gammaDist(2, 2), green, 0.99))

	doc.Note(`These plots were easy to generate because the
		 distributions types, provided by gonum/stat/distuv,
		 each provide analytical methods for the Mean,
		 Mode, Probability, CDF, and Quantile functions.
		 Having these methods given to us, makes it easy to
		 set the plot's X and Y axis ranges.`)
}

func noteDisplaying(doc essay.Document) {

	doc.Note(`Often, we are given data that we suspect was "drawn"
		 from a probability distribution, that we wish to
		 estimate or convey to a user.  For a graphical
		 display, we can draw a number of single-dimensional
		 values and display them on the number line.  The
		 intuition is that the density of points on the line
		 is proportional with the variable's probability at that
		 location.`)

	doc.Note(`Here are 100 values drawn from the standard normal
		distribution.  This is not a great way to visualize
		these data.`,
		plot1DExp(normalDist(0, 1), red, 0.99, 100))

	doc.Note(`It especially doesn't work when there are too many
		values.  Here's how it looks with 1000 values.  Note
		the X-axis is stretched because of outliers in the data.`,
		plot1DExp(normalDist(0, 1), red, 0.99, 1000))
}

func noteHistograms(doc essay.Document) {

	doc.Note(`Histograms offer a better way to display the
	distribution of a single-dimensional random variable.
	Histograms divide the supported values into discrete ranges,
	referred to as bins, and display the count of values that fall
	into each bin.  When you normalize the counts observed in the
	bins to 1, you have an empirical distribution.`)

	doc.Note(`The number of bins in the histogram should be
	significantly less than the number of data points, or it
	suffers at visualization just as when we displayed the points
	directly on a number-line (each bin will have a count of one).`)

	doc.Note(`Let's look at 100 points in 20 bins. In the
	histograms presented here, bins are equally-sized and span from the
	minimum to the maximum observed value.`,
		plotHist(normalDist(0, 1), red, 0.99, 20, 100))

	doc.Note(`Supplied with enough data, the empirical
	distribution of a normalized histogram should converge with
	the probability distribution from which it was drawn. Here are
	10K values in 20 bins.`, plotHist(normalDist(0, 1), red, 0.99, 20,
		10000))

	doc.Note(`10K values with 100 bins.`,
		plotHist(normalDist(0, 1), red, 0.99, 100, 10000))

	doc.Note(`10K values with 1000 bins.`,
		plotHist(normalDist(0, 1), red, 0.99, 1000, 10000))

	doc.Note(`The above looks clearly worse than the preceding histogram.  Let's look at several rules for choosing histogram size.`)

	doc.Section(`Histogram Bins & Count`, func(doc essay.Document) {
		doc.Note(`Here are several well-known rules for sizing the histogram 
			based on the sample size and (maybe) distribution, shown here
			for the standard normal distribution.`,
			showHistSizes(doc, normalDist(0, 1)))
	})
}

func noteSkewed(doc essay.Document) {

	unitExp := exponentialDist(1)
	data := drawFrom(unitExp, 1000)

	doc.Note(`Skewed distributions technically are those whose median is distant 
		from the mean, which generally indicates there is a high-density probability region
		offset from a low-density probability region.  Here are 1000 samples from the unit
		exponential distribution.`,
		plotHist(exponentialDist(1), blue, 0.99, 20, 1000),

		`Since we have an exponential distribution, it's natural to look at the Y-axis
                in log-scale, but it raises questions in the context of a histogram.  Whereas
		a log-scale plot is good at displaying relative frequencies, it is not terribly
		good at displaying absolute frequenceis because log(0) is undefined.  Histograms 
		typically must display zeros, but efforts to correct for zero distort the data.  
		Consider the bin with the smallest non-zero count.  There is a log-linear 
		relationship between the non-zero-frequency data points, but the height of the
		bar representing the smallest non-zero count is arbitrary.  Here are three 
		log-scale histograms representing the same data, for illustration.`,
		essay.RowTable(
			plotLogHistData(unitExp, blue, 0.99, 20, 0.01, data),
			plotLogHistData(unitExp, blue, 0.99, 20, 0.1, data),
			plotLogHistData(unitExp, blue, 0.99, 20, 1, data),
		),

		`Using log-scale on a histogram effectively gives the bins variable size, but
		it doesn't remove the question of how many bins to use; they are independent 
		questions.`,

		`Another skewed distribution is the log-normal.`,
		// @@@ This doesn't display well, trouble w/ X-max and Y-max.

		essay.RowTable(
			plot1(logNormalDist(1, 1), red, 0.99),
			plot1(logNormalDist(5, 5), red, 0.99),
			plot1(logNormalDist(10, 10), red, 0.99)),
	)

	doc.Section(`Exponential Distribution: Bin Count`, func(doc essay.Document) {

		doc.Note(`The exponential distribution is not very
			skewed, as skewed distributions go, so let's
			forget about the issue of log-scale for
			now. Here are several choices for histogram
			size using the exponential distribution, with
			linear scale.`,
			showHistSizes(doc, unitExp))
	})
}

func noteHeavyTail(doc essay.Document) {
	pdist := paretoDist(1, 1)
	data := drawFrom(pdist, 10000)

	doc.Note(`Heavy-tail distributions can be defined in various ways, but are
		generally extremely skewed distributions which nevertheless happen
		in nature.  Sometimes these are called extreme-value distributions
		and there are alternate terms such as "long-tailed" and "fat-tailed" 
		to distinguish sub-categories.  Power-law distributions are heavy-
		tailed (e.g., Zipf).`,

		`The thing is, these distributions are much harder to
		work with, so are not as simple to plot.  For example,
		they may define a Quantile function, which we've been
		using so far to set minimum coverage for the X-axis
		range.`,

		`What does it mean, to have infinite mean?  It means
		as we sample from the distribution, the mean continues
		to rise as the probability falls.  The Pareto distribution is a power law
		distribution for parameter α > 0 with infinite mean
		when α <= 1.  We'll look at the Pareto distribution
		with parameters x_min = 1 and α = 1, setting the X
		range and log-scale Y range based on 1K, 100K, and 10M
		samples.`,

		essay.HeadRowTable(
			"1K Samples",
			plotHeavyCDF1Log(pdist, blue, 1e3),
			"100K Samples",
			plotHeavyCDF1Log(pdist, blue, 1e5),
			"10M Samples",
			plotHeavyCDF1Log(pdist, blue, 1e7),
		),

		`As the number of samples rises, the shape of the
		curve changes because see more of it.  This is in
		contrast to the normal and exponential distributions,
		which have scale-invariant shape.  Again, when
		displaying log-scale data, we decide the height of the
		smallest bin.  These are three histograms of the same
		data, in 100 bins.`,

		essay.RowTable(
			plotHeavyData1Log(pdist, red, 100, 0.01, data),
			plotHeavyData1Log(pdist, red, 100, 0.1, data),
			plotHeavyData1Log(pdist, red, 100, 1, data),
		),

		`The natural way to visualize Pareto distribution is
		on a double-log scale. `,

		essay.HeadRowTable(
			"1K Samples",
			plotHeavyCDF(pdist, blue, 1e3),
			"100K Samples",
			plotHeavyCDF(pdist, blue, 1e5),
			"10M Samples",
			plotHeavyCDF(pdist, blue, 1e7),
		),

		`To test for a hypothesized Pareto distribution, plot
		the samples on a double-log scale.  This is the same
		data as shown in the histograms above, using a log-log
		histogram.`,

		essay.RowTable(
			plotHeavyData(pdist, green, 100, 0.01, data),
			plotHeavyData(pdist, green, 100, 0.1, data),
			plotHeavyData(pdist, green, 100, 1, data),
		),
	)
}

func plotHeavyCDF(dist HeavyDist, col color.Color, samples int) essay.Renderer {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pfunc := plotter.NewFunction(dist.Prob)
	pfunc.Color = col
	pfunc.Width = vg.Points(1)
	pfunc.Samples = 1000

	p.Title.Text = dist.String()
	p.Title.Padding = vg.Points(5)

	p.X.Min = 1
	p.X.Max = 1
	p.Y.Min = 0
	p.Y.Max = 1

	var sampleData XY1D
	for i := 0; i < samples; i++ {
		x := dist.Rand()
		p.X.Max = math.Max(p.X.Max, x)
		sampleData = append(sampleData, x)
	}

	setLogScale(p, Y, dist, 1, sampleData)
	setLogScale(p, X, dist, 1, sampleData)

	p.Add(plotter.NewGrid(), pfunc)

	return num.Plot(p, 300, 300)
}

func plotHeavyCDF1Log(dist HeavyDist, col color.Color, samples int) essay.Renderer {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pfunc := plotter.NewFunction(dist.Prob)
	pfunc.Color = col
	pfunc.Width = vg.Points(1)
	pfunc.Samples = 1000

	p.Title.Text = dist.String()
	p.Title.Padding = vg.Points(5)

	p.X.Min = 1
	p.X.Max = 1
	p.Y.Min = 0
	p.Y.Max = 1

	var sampleData XY1D
	for i := 0; i < samples; i++ {
		x := dist.Rand()
		p.X.Max = math.Max(p.X.Max, x)
		sampleData = append(sampleData, x)
	}

	setLogScale(p, Y, dist, 1, sampleData)

	p.Add(plotter.NewGrid(), pfunc)

	return num.Plot(p, 300, 300)
}

func plotHeavyData1Log(dist HeavyDist, col color.Color, bins int, factor float64, data XY1D) essay.Renderer {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	s, err := plotter.NewHist(data, bins)
	s.FillColor = col
	s.LineStyle.Width = vg.Points(0.1)
	if err != nil {
		panic(err)
	}
	s.Normalize(1)

	pfunc := plotter.NewFunction(dist.Prob)
	pfunc.Color = black
	pfunc.Width = vg.Points(1)
	pfunc.Samples = 1000

	p.Title.Padding = vg.Points(5)
	p.Title.Text = dist.String()

	setLogScale(p, Y, dist, factor, data)

	p.Add(s, pfunc)

	return num.Plot(p, 300, 300)
}

func plotHeavyData(dist HeavyDist, col color.Color, bins int, factor float64, data XY1D) essay.Renderer {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	s, err := loghist.NewHist(data, loghist.LinearTransformer{}, loghist.FixedBinner(bins))
	s.FillColor = col
	s.LineStyle.Width = vg.Points(0.5)
	if err != nil {
		panic(err)
	}
	s.Normalize(1)

	pfunc := plotter.NewFunction(dist.Prob)
	pfunc.Color = black
	pfunc.Width = vg.Points(1)
	pfunc.Samples = 1000

	p.Title.Padding = vg.Points(5)
	p.Title.Text = dist.String()

	setLogScale(p, Y, dist, factor, data)
	setLogScale(p, X, dist, factor, data)

	p.Add(s, pfunc)

	return num.Plot(p, 300, 300)
}

func setLogScale(p *plot.Plot, ax axis, dist HeavyDist, factor float64, data XY1D) {
	var (
		smallest = math.Inf(+1)
		biggest  = smallest
	)
	for _, d := range data {
		if ax == X {
			if d < smallest {
				smallest = d
			}
		} else if p := dist.Prob(d); p < smallest {
			smallest = p
		}

	}
	smallest *= factor

	if ax == X {
		p.X.Scale = num.ClipScale{smallest, biggest, plot.LogScale{}}
		p.X.Tick.Marker = num.ClipTicker{smallest, biggest, plot.LogTicks{}}
	} else {
		p.Y.Scale = num.ClipScale{smallest, biggest, plot.LogScale{}}
		p.Y.Tick.Marker = num.ClipTicker{smallest, biggest, plot.LogTicks{}}
	}
}

func showHistSizes(doc essay.Document, dist Dist) essay.Table {
	// Note https://en.wikipedia.org/wiki/Histogram#Number_of_bins_and_width

	var cells [][]interface{}
	left := []interface{}{
		100,
		500,
		1000,
		5000,
		10000,
		50000,
		100000,
	}

	algs := []struct {
		N string
		F func([]float64) float64
	}{
		{
			"Square-root ~x^1/2",
			func(x []float64) float64 {
				return math.Sqrt(float64(len(x)))
			},
		},
		{
			"Rice ~x^1/3",
			func(x []float64) float64 {
				y := math.Pow(float64(len(x)), 1.0/3.0)
				return math.Ceil(2 * y)
			},
		},
		{
			"Sturges ~log2(x)",
			func(x []float64) float64 {
				return math.Ceil(math.Log2(float64(len(x)))) + 1
			},
		},
	}
	// Note: there are also a number of rules better suited for
	// skewed distributions: "Doane", "Scott", "Freedman–Diaconis"
	// rules, then a couple of size-choice algorithms based on an
	// optimization: "Min-Squared-Error", "Shimazaki and Shinomoto".

	// Note: gonum/plot default is the square root of the sum of
	// the Y values.

	top := []interface{}{
		"Sample size",
	}

	for _, alg := range algs {
		top = append(top, alg.N)
	}

	for _, samplesi := range left {
		samples := samplesi.(int)
		row := []interface{}{}
		data := drawFrom(dist, samples)
		for _, alg := range algs {
			histSize := alg.F(data)
			cell, _ := plotHistData(dist, greys, .99, int(math.Ceil(histSize)), data)
			row = append(row, cell)
		}
		cells = append(cells, row)
	}
	return essay.Table{
		Cells:   cells,
		TopRow:  top,
		LeftCol: left,
	}
}

// doc.Section("TODO: Log-Linear Sizing", "We commonly use log-linear sized bins.")
// doc.Section("TODO: Variable-width Sizing", "There are variable-sized-bin algorithms like Q-digest and T-digest.")

func plotLogHist(dist Dist, col color.Color, show float64, bins int, factor float64, cnt int) essay.Renderer {
	return plotLogHistData(dist, col, show, bins, factor, drawFrom(dist, cnt))
}

func plotLogHistData(dist Dist, col color.Color, show float64, bins int, factor float64, data XY1D) essay.Renderer {
	pdata, p := plotHistData(dist, col, show, bins, data)

	setLogScale(p, Y, dist, factor, data)

	return pdata
}

func plotHist(dist Dist, col color.Color, show float64, bins int, cnt int) essay.Renderer {
	data, _ := plotHistData(dist, col, show, bins, drawFrom(dist, cnt))
	return data
}

func plotHistData(dist Dist, col color.Color, show float64, bins int, data XY1D) (essay.Renderer, *plot.Plot) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	s, err := plotter.NewHist(data, bins)
	s.FillColor = col
	s.LineStyle.Width = vg.Points(0.1)
	if err != nil {
		panic(err)
	}
	s.Normalize(1)

	pfunc := plotter.NewFunction(dist.Prob)
	pfunc.Color = black
	pfunc.Width = vg.Points(1)
	pfunc.Samples = 1000

	p.Title.Padding = vg.Points(5)
	p.Title.Text = dist.String()
	p.X.Min = dist.Quantile(1 - show)
	p.X.Max = dist.Quantile(show)
	p.Y.Min = 0
	p.Y.Max = dist.Prob(dist.Mode())

	p.Add(s, pfunc)

	return num.Plot(p, 300, 150), p
}

func plot1DExp(dist Dist, col color.Color, show float64, cnt int) essay.Renderer {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	vs := make(XY1D, cnt)
	for i := 0; i < cnt; i++ {
		vs[i] = dist.Rand()
	}

	s, err := plotter.NewScatter(vs)
	if err != nil {
		panic(err)
	}
	s.GlyphStyle.Color = col
	s.GlyphStyle.Radius = vg.Points(40)
	s.GlyphStyle.Shape = LineGlyph{}

	pfunc := plotter.NewFunction(dist.Prob)
	pfunc.Color = greya
	pfunc.Width = vg.Points(1)
	pfunc.Samples = 1000

	p.Title.Padding = vg.Points(5)
	p.Title.Text = dist.String()
	p.X.Min = dist.Quantile(1 - show)
	p.X.Max = dist.Quantile(show)
	p.Y.Min = 0
	p.Y.Max = dist.Prob(dist.Mode())

	p.Add(pfunc, s)

	return num.Plot(p, 500, 200)
}

func plotCDF(dist Dist, col color.Color, show float64) essay.Renderer {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pfunc := plotter.NewFunction(dist.Prob)
	pfunc.Color = col
	pfunc.Width = vg.Points(1)
	pfunc.Samples = 1000

	cfunc := plotter.NewFunction(dist.CDF)
	cfunc.Color = greya
	cfunc.Width = vg.Points(1)
	cfunc.Samples = 1000

	p.Title.Text = dist.String()
	p.Title.Padding = vg.Points(5)
	p.X.Min = dist.Quantile(1 - show)
	p.X.Max = dist.Quantile(show)
	p.Y.Min = 0
	p.Y.Max = 1

	p.Legend.Top = true
	p.Legend.Left = true
	p.Legend.Add("CDF", cfunc)
	p.Legend.Add("PDF", pfunc)

	p.Add(plotter.NewGrid(), pfunc, cfunc)

	return num.Plot(p, 300, 300)
}

func plot1(dist Dist, color color.Color, show float64) essay.Renderer {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pfunc := plotter.NewFunction(dist.Prob)
	pfunc.Color = color
	pfunc.Width = vg.Points(1)
	pfunc.Samples = 1000

	p.Title.Text = dist.String()
	p.Title.Padding = vg.Points(5)
	p.X.Min = dist.Quantile(1 - show)
	p.X.Max = dist.Quantile(show)
	p.Y.Min = 0
	p.Y.Max = dist.Prob(dist.Mode())

	p.Add(plotter.NewGrid(), pfunc)

	return num.Plot(p, 300, 300)
}

func plot2(title string, dist1, dist2 Dist, c1, c2 color.Color, show float64) essay.Renderer {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	pfunc1 := plotter.NewFunction(dist1.Prob)
	pfunc1.Color = c1
	pfunc1.Width = vg.Points(1)
	pfunc1.Samples = 1000

	pfunc2 := plotter.NewFunction(dist2.Prob)
	pfunc2.Color = c2
	pfunc2.Width = vg.Points(1)
	pfunc2.Samples = 1000

	p.Title.Text = title
	p.Title.Padding = vg.Points(5)
	p.X.Min = math.Min(dist1.Quantile(1-show), dist2.Quantile(1-show))
	p.X.Max = math.Max(dist1.Quantile(show), dist2.Quantile(show))
	p.Y.Min = 0
	p.Y.Max = math.Max(dist1.Prob(dist1.Mode()), dist2.Prob(dist2.Mode()))

	p.Add(plotter.NewGrid(), pfunc1, pfunc2)

	return num.Plot(p, 300, 300)
}

func (g gonumDist) String() string {
	return g.name
}

func (g heavyDist) String() string {
	return g.name
}

func normalDist(mu, sigma float64) Dist {
	return gonumDist{
		GoodDistAPI: distuv.Normal{
			Mu:    mu,
			Sigma: sigma,
		},
		name: fmt.Sprintf("Normal μ=%.1f σ=%.1f", mu, sigma),
	}
}

func betaDist(alpha, beta float64) Dist {
	return gonumDist{
		GoodDistAPI: distuv.Beta{
			Alpha: alpha,
			Beta:  beta,
		},
		name: fmt.Sprintf("Beta α=%.1f β=%.1f", alpha, beta),
	}
}

func exponentialDist(rate float64) Dist {
	return gonumDist{
		GoodDistAPI: distuv.Exponential{
			Rate: rate,
		},
		name: fmt.Sprintf("Exponential λ=%.1f", rate),
	}
}

func gammaDist(alpha, beta float64) Dist {
	return gonumDist{
		GoodDistAPI: distuv.Gamma{
			Alpha: alpha, // Shape
			Beta:  beta,  // Rate
		},
		name: fmt.Sprintf("Gamma α=%.1f β=%.1f", alpha, beta),
	}
}

func paretoDist(xm, alpha float64) HeavyDist {
	return heavyDist{
		DistAPI: distuv.Pareto{
			Xm:    xm,    // (Minimum-Value) Scale
			Alpha: alpha, // Shape
		},
		name: fmt.Sprintf("Pareto Mx=%.1f α=%.1f", xm, alpha),
	}
}

func logNormalDist(mu, sigma float64) Dist {
	return gonumDist{
		GoodDistAPI: distuv.LogNormal{
			Mu:    mu,
			Sigma: sigma,
		},
		name: fmt.Sprintf("LogNormal μ=%.1f σ=%.1f", mu, sigma),
	}
}

func drawFrom(dist DistAPI, cnt int) []float64 {
	vs := make(XY1D, cnt)
	for i := 0; i < cnt; i++ {
		vs[i] = dist.Rand()
	}
	return vs
}

func (x XY1D) Len() int {
	return len(x)
}

func (d XY1D) XY(i int) (float64, float64) {
	// @@@ NOTNOT NOT
	return d[i], 0
}

func (d XY1D) Value(i int) float64 {
	return d[i]
}

func (LineGlyph) DrawGlyph(c *draw.Canvas, sty draw.GlyphStyle, pt vg.Point) {
	c.SetLineStyle(draw.LineStyle{Color: sty.Color, Width: vg.Points(0.3)})
	r := sty.Radius
	var p vg.Path
	p.Move(vg.Point{X: pt.X, Y: pt.Y + 2*r})
	p.Line(vg.Point{X: pt.X, Y: pt.Y})
	c.Stroke(p)
}
