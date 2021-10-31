package main

import (
	"fmt"
	"image/color"
	"runtime"

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
	doc.Note(`This essay uses gonum/plot and gonum/stat/distuv to
		display CDFs, PDFs, histograms, and samples drawn from
		univariate distributions.`)

	doc.Section("Probability Distributions", noteDistributions)
}

func noteDistributions(doc essay.Document) {

	doc.Note(`Kolmo!!`, plot1(red, 0.99))

}
func plot1(color color.Color, show float64) essay.Renderer {
	p := plot.New()

	pfunc := plotter.NewFunction(func(x float64) float64 {
		defer func() {
			if err := recover(); err != nil {
				s := make([]byte, 4096)
				s = s[0:runtime.Stack(s, false)]
				fmt.Println("ERR", err, "\n", string(s))
			}
		}()
		return kolmogorov.K(1000, x)
	})
	pfunc.Color = color
	pfunc.Width = vg.Points(1)
	pfunc.Samples = 100

	p.Title.Text = "kolmogorov D(1000)"
	p.Title.Padding = vg.Points(5)

	// p.X.Min = dist.Quantile(1 - show)
	// p.X.Max = dist.Quantile(show)
	// p.Y.Min = 0
	// p.Y.Max = dist.Prob(dist.Mode())

	p.Add(plotter.NewGrid(), pfunc)

	return num.Plot(p, 300, 300)
}
