package num

import (
	"image/color"
	"math"

	"github.com/lightstep/sandbox/jmacd/gonum/loghist"
	"gonum.org/v1/plot"
)

type (
	HBuilder struct {
		bins       []loghist.HistogramBin
		trans      loghist.DataTransformer
		normalize  bool
		centroids  bool
		fillColors []color.Color
		lineColor  color.Color
		name       string
	}

	nonzeroRange struct {
		*loghist.Histogram
	}
)

func NewHistogram() HBuilder {
	return HBuilder{}
}

func (h HBuilder) Transformer(trans loghist.DataTransformer) HBuilder {
	h.trans = trans
	return h
}

func (h HBuilder) Bins(bins []loghist.HistogramBin) HBuilder {
	h.bins = bins
	return h
}

func (h HBuilder) Normalize() HBuilder {
	h.normalize = true
	return h
}

func (h HBuilder) Centroids() HBuilder {
	h.centroids = true
	return h
}

func (h HBuilder) FillColors(colors []color.Color) HBuilder {
	h.fillColors = colors
	return h
}

func (h HBuilder) LineColor(color color.Color) HBuilder {
	h.lineColor = color
	return h
}

func (h HBuilder) Name(name string) HBuilder {
	h.name = name
	return h
}

func (h HBuilder) addTo(builder Builder) (thumb plot.Thumbnailer, valranger, plotranger plot.DataRanger, err error) {
	hist := loghist.NewHistogram(h.trans, h.bins)

	if h.normalize {
		hist.Normalize(1)
	}
	if h.centroids {
		hist.Centroids()
	}
	if h.fillColors != nil {
		hist.FillColors = h.fillColors
	}
	if h.lineColor != nil {
		hist.LineStyle.Color = h.lineColor
	}
	builder.Plot.Add(hist)
	return hist, nonzeroRange{hist}, hist, nil
}

func (h nonzeroRange) DataRange() (xmin, xmax, ymin, ymax float64) {
	xmin = math.Inf(+1)
	xmax = math.Inf(-1)
	ymin = math.Inf(+1)
	ymax = math.Inf(-1)
	for _, bin := range h.Bins {
		if bin.Max > xmax {
			xmax = bin.Max
		}
		if bin.Min < xmin {
			xmin = bin.Min
		}
		if bin.Weight > ymax {
			ymax = bin.Weight
		}
		if bin.Weight > 0 && bin.Weight < ymin {
			ymin = bin.Weight
		}
	}
	return
}

func (h HBuilder) namePlot() string {
	return h.name
}
