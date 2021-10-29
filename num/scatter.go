package num

import (
	"image/color"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var (
	defaultScatterRadius = vg.Points(1)
	defaultScatterColor  = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	defaultScatterShape  = draw.CircleGlyph{}
)

type (
	SBuilder struct {
		data   plotter.XYer
		radius vg.Length
		color  color.Color
		shape  draw.GlyphDrawer
		style  func(int) draw.GlyphStyle
		name   string
	}
)

func Scatter(data plotter.XYer) SBuilder {
	return SBuilder{
		data:   data,
		radius: defaultScatterRadius,
		color:  defaultScatterColor,
		shape:  defaultScatterShape,
	}
}

func (s SBuilder) Radius(points float64) SBuilder {
	s.radius = vg.Points(points)
	return s
}

func (s SBuilder) Shape(shape draw.GlyphDrawer) SBuilder {
	if shape != nil {
		s.shape = shape
	}
	return s
}

func (s SBuilder) Color(color color.Color) SBuilder {
	if color != nil {
		s.color = color
	}
	return s
}

func (s SBuilder) Style(style func(int) draw.GlyphStyle) SBuilder {
	s.style = style
	return s
}

func (s SBuilder) Name(name string) SBuilder {
	s.name = name
	return s
}

func (s SBuilder) addTo(builder Builder) (thumb plot.Thumbnailer, valranger, plotranger plot.DataRanger, err error) {
	scatter, err := plotter.NewScatter(s.data)
	if err != nil {
		return nil, nil, nil, err
	}

	if s.style != nil {
		scatter.GlyphStyleFunc = func(i int) draw.GlyphStyle {
			gs := s.style(i)
			if gs.Color == nil {
				gs.Color = s.color
			}
			if gs.Shape == nil {
				gs.Shape = s.shape
			}
			if gs.Radius == 0 {
				gs.Radius = s.radius
			}
			return gs
		}
	} else {
		scatter.GlyphStyle.Shape = s.shape
		scatter.GlyphStyle.Radius = s.radius
		scatter.GlyphStyle.Color = s.color
	}
	builder.Plot.Add(scatter)
	return scatter, scatter, nil, nil
}

func (s SBuilder) namePlot() string {
	return s.name
}
