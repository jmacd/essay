package main

import (
	"image/color"

	"github.com/jmacd/essay"
	"github.com/jmacd/essay/examples/kolmogorov/kolmogorov"
	"github.com/jmacd/essay/num"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
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
)

func main() {
	essay.Main("Kolmogorov distribution", write)
}

func write(doc essay.Document) {
	doc.Note(`Let's see.`)

	doc.Section("Probability Distributions", noteDistributions)
}

func noteDistributions(doc essay.Document) {

	doc.Note(`Kolmo!!`, plot1(red, 0.99))

}

func plot1(color color.Color, show float64) essay.Renderer {
	// const testN = 10
	const many = 1000

	p := plot.New()

	// cdf := plotter.NewFunction(func(x float64) float64 {
	// 	return kolmogorov.K(testN, x)
	// })
	// cdf.Color = color
	// cdf.Width = vg.Points(1)
	// cdf.Samples = many

	p.Title.Text = "kolmogorov D"
	p.Title.Padding = vg.Points(5)

	p.X.Min = 0
	p.X.Max = 1
	p.Y.Min = 0
	p.Y.Max = 7

	p.Add(plotter.NewGrid())

	for testN := 2; testN <= 10; testN += 1 {
		testN := testN
		pdf := plotter.NewFunction(func(x float64) float64 {
			const epsilon = 1e-7
			return (kolmogorov.K(testN, x) - kolmogorov.K(testN, x-epsilon)) / epsilon
		})
		if testN == 10 {
			pdf.Color = blue
		} else if testN == 100 {
			pdf.Color = red
		} else {
			pdf.Color = black
		}
		pdf.Width = vg.Points(1)
		pdf.Samples = many

		p.Add(pdf)
	}

	return num.Plot(p, 800, 800)
}
