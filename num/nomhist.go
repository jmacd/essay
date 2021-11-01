package num

import (
	"image/color"

	"github.com/jmacd/essay/lib/gonum/nomhist"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
)

type (
	NomBuilder struct {
		valuer plotter.Valuer
		colorf func(int) color.Color
	}
)

func NominalHistogram(valuer plotter.Valuer, colorf func(int) color.Color) NomBuilder {
	return NomBuilder{
		valuer: valuer,
		colorf: colorf,
	}
}

func (n NomBuilder) addTo(builder Builder) (thumb plot.Thumbnailer, valranger, plotranger plot.DataRanger, err error) {
	nh, err := nomhist.NewNominalHistogram(n.valuer, n.colorf)
	if err != nil {
		panic(err)
	}

	builder.Plot.Add(nh)
	return nil, nh, nh, nil
}

func (f NomBuilder) namePlot() string {
	return ""
}
