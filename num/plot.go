package num

import (
	"bytes"
	"image"
	"math"

	"github.com/jmacd/essay"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const (
	defaultWidth  = 300
	defaultHeight = 300

	axisWidth = 1

	// The baseline for a log-scale plot will be set so that
	// distance bewteen the baseline and the minimum value is the
	// reciprocal of this factor times the distance between the
	// maximum value and the mimimum value.  A larger factor
	// places relatively more visual emphasis on smaller values,
	// while compressing the upper end of the log-scaled range.
	defaultLogBaseFactor = 8
)

var (
	defaultLegendPadding = vg.Points(5)
)

type (
	Builder struct {
		Plot         *plot.Plot
		Width        int
		Height       int
		Plotters     []Plotter
		DrawLegend   bool
		XAxis, YAxis ABuilder
		Ranger       Ranger
	}

	Ranger interface {
		DataRange() (xmin, xmax, ymin, ymax float64)
	}

	Plotter interface {
		namePlot() string
		addTo(Builder) (p plot.Thumbnailer, v, r plot.DataRanger, e error)
	}
)

func (builder Builder) Render(builtin essay.Builtin) (interface{}, error) {
	img := builder.Image(essay.PNG)
	return builtin.RenderImage(img)
}

func (builder Builder) setupPlot() {
	minx, miny := math.Inf(+1), math.Inf(+1)
	maxx, maxy := math.Inf(-1), math.Inf(-1)

	for _, p := range builder.Plotters {
		thumbnailer, valranger, plotranger, err := p.addTo(builder)
		if err != nil {
			panic(err)
		}

		if valranger != nil {
			xmin, xmax, ymin, ymax := valranger.DataRange()

			minx = math.Min(minx, xmin)
			miny = math.Min(miny, ymin)
			maxx = math.Max(maxx, xmax)
			maxy = math.Max(maxy, ymax)
		}

		if plotranger != nil {
			xmin, xmax, ymin, ymax := plotranger.DataRange()

			builder.Plot.X.Min = math.Min(builder.Plot.X.Min, xmin)
			builder.Plot.X.Max = math.Min(builder.Plot.X.Max, xmax)
			builder.Plot.Y.Min = math.Min(builder.Plot.Y.Min, ymin)
			builder.Plot.Y.Max = math.Min(builder.Plot.Y.Max, ymax)
		}

		if builder.DrawLegend && thumbnailer != nil {
			// TODO This needs more flexibility
			builder.Plot.Legend.Add(p.namePlot(), thumbnailer)
			builder.Plot.Legend.Padding = defaultLegendPadding
			builder.Plot.Legend.Top = true
			builder.Plot.Legend.YOffs = defaultLegendPadding
			builder.Plot.Legend.XOffs = -defaultLegendPadding
		}
	}

	if builder.XAxis.nominal != nil {
		builder.Plot.NominalX(builder.XAxis.nominal...)
	}
	if builder.XAxis.hide {
		// builder.Plot.HideX()
		builder.Plot.X.Tick.Length = 0
		builder.Plot.X.Width = axisWidth
		builder.Plot.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{})
		builder.Plot.X.Label.Text = builder.Plot.Title.Text
		builder.Plot.X.Label.TextStyle.Font.Size = 18
		builder.Plot.X.Label.TextStyle.XAlign = draw.XCenter
		builder.Plot.Title.Text = ""
	}

	if builder.XAxis.logscale {
		factor := builder.XAxis.factor
		base := math.Pow(math.Pow(minx, factor+1)/maxx, 1.0/factor)
		builder.Plot.X.Scale = MinLogScale(base)
		builder.Plot.X.Tick.Marker = MinLogTicks(base)
	}
	if builder.YAxis.logscale {
		factor := builder.YAxis.factor
		base := math.Pow(math.Pow(miny, factor+1)/maxy, 1.0/factor)
		builder.Plot.Y.Scale = MinLogScale(base)
		builder.Plot.Y.Tick.Marker = MinLogTicks(base)
	}

	if builder.XAxis.min != nil {
		builder.Plot.X.Min = *builder.XAxis.min
	}
	if builder.XAxis.max != nil {
		builder.Plot.X.Max = *builder.XAxis.max
	}
	if builder.YAxis.min != nil {
		builder.Plot.Y.Min = *builder.YAxis.min
	}
	if builder.YAxis.max != nil {
		builder.Plot.Y.Max = *builder.YAxis.max
	}
	if builder.YAxis.hide {
		builder.Plot.HideY()
		// builder.Plot.Y.Tick.Length = 0
		// builder.Plot.Y.Width = axisWidth
		// builder.Plot.Y.Tick.Marker = plot.ConstantTicks([]plot.Tick{})

	}

	if builder.Ranger != nil {
		xmin, xmax, ymin, ymax := builder.Ranger.DataRange()
		builder.Plot.X.Min = xmin
		builder.Plot.X.Max = xmax
		builder.Plot.Y.Min = ymin
		builder.Plot.Y.Max = ymax
	}
}

func (builder Builder) Image(kind essay.ImageKind) essay.EncodedImage {
	builder.setupPlot()

	w := vg.Length(builder.Width)
	h := vg.Length(builder.Height)

	var buf bytes.Buffer
	if writer, err := builder.Plot.WriterTo(w, h, string(kind)); err != nil {
		panic(err)
	} else if _, err := writer.WriteTo(&buf); err != nil {
		panic(err)
	}
	return essay.EncodedImage{
		Kind: kind,
		Bounds: image.Rectangle{
			Min: image.ZP,
			Max: image.Pt(builder.Width, builder.Height),
		},
		Data: buf.Bytes(),
	}
}

func Plot(p *plot.Plot, w, h int) Builder {
	return Builder{
		Plot:   p,
		Width:  w,
		Height: h,
	}
}

func NewPlot() Builder {
	p := plot.New()
	return Builder{
		Plot:   p,
		Width:  defaultWidth,
		Height: defaultHeight,
	}
}

func (b Builder) Title(t string) Builder {
	b.Plot.Title.Text = t
	// b.Plot.Title.TextStyle.Font.Size = 24
	// b.Plot.Title.TextStyle.XAlign = draw.XCenter
	b.Plot.Title.Padding = vg.Points(5)
	return b
}

func (b Builder) Size(w, h int) Builder {
	b.Width = w
	b.Height = h
	return b
}

func (b Builder) Range(r Ranger) Builder {
	// TODO want to set the X range only, for function plots, to
	// accomodate the min/max observed Y given the X range only.
	b.Ranger = r
	return b
}

func (b Builder) X(a ABuilder) Builder {
	b.XAxis = a
	return b
}

func (b Builder) Y(a ABuilder) Builder {
	b.YAxis = a
	return b
}

func (b Builder) Add(p ...Plotter) Builder {
	b.Plotters = append(b.Plotters, p...)
	return b
}

func (b Builder) AddEach(n int, f func(i int) Plotter) Builder {
	for i := 0; i < n; i++ {
		b = b.Add(f(i))
	}
	return b
}

func (b Builder) Legend() Builder {
	b.DrawLegend = true
	return b
}
