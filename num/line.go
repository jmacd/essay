package num

import (
	"image/color"

	"github.com/jmacd/gospline"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

const fsamples = 1000

var (
	defaultLineWidth = vg.Points(1)
	defaultLineColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}
)

type (
	XYoker interface {
		XYok(int) (x, y float64, ok bool)
	}

	LBuilder struct {
		data   plotter.XYer
		width  vg.Length
		color  color.Color
		name   string
		smooth bool
	}
)

func Line(data plotter.XYer) LBuilder {
	return LBuilder{
		data:  data,
		width: defaultLineWidth,
		color: defaultLineColor,
	}
}

func (l LBuilder) Width(points float64) LBuilder {
	l.width = vg.Points(points)
	return l
}

func (l LBuilder) Color(color color.Color) LBuilder {
	l.color = color
	return l
}

func (l LBuilder) Name(name string) LBuilder {
	l.name = name
	return l
}

func (l LBuilder) Smooth() LBuilder {
	l.smooth = true
	return l
}

func (l LBuilder) addTo(builder Builder) (thumb plot.Thumbnailer, valranger, plotranger plot.DataRanger, err error) {
	if !l.smooth || l.data.Len() == 0 {
		line, err := plotter.NewLine(l.data)
		if err != nil {
			return nil, nil, nil, err
		}
		line.LineStyle.Width = l.width
		line.LineStyle.Color = l.color
		builder.Plot.Add(line)
		return line, line, nil, nil
	}

	var xs []float64
	var ys []float64

	for i := 0; i < l.data.Len(); i++ {
		x, y, _ := l.data.(XYoker).XYok(i)
		xs = append(xs, x)
		ys = append(ys, y)
	}

	sp := gospline.NewMonotoneSpline(xs, ys)

	line := plotter.NewFunction(func(x float64) float64 {
		return sp.At(x)
	})
	line.Samples = fsamples
	line.LineStyle.Width = l.width
	line.LineStyle.Color = l.color
	builder.Plot.Add(line)
	return line, nil, nil, nil
}

func (l LBuilder) namePlot() string {
	return l.name
}
