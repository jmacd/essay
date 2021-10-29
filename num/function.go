package num

import (
	"image/color"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type (
	FBuilder struct {
		F       func(float64) float64
		width   vg.Length
		color   color.Color
		samples int
		name    string
	}
)

func Function(F func(float64) float64) FBuilder {
	return FBuilder{
		F:     F,
		width: defaultLineWidth,
		color: defaultLineColor,
	}
}

func (f FBuilder) Width(points float64) FBuilder {
	f.width = vg.Points(points)
	return f
}

func (f FBuilder) Color(color color.Color) FBuilder {
	f.color = color
	return f
}

func (f FBuilder) Name(name string) FBuilder {
	f.name = name
	return f
}

func (f FBuilder) Samples(samples int) FBuilder {
	f.samples = samples
	return f
}

func (f FBuilder) addTo(builder Builder) (thumb plot.Thumbnailer, valranger, plotranger plot.DataRanger, err error) {
	fun := plotter.NewFunction(f.F)
	fun.LineStyle.Width = f.width
	fun.LineStyle.Color = f.color
	fun.Samples = f.samples
	builder.Plot.Add(fun)
	return fun, nil, nil, nil
}

func (f FBuilder) namePlot() string {
	return f.name
}
