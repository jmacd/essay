package plot

import (
	"math/rand"

	ms "github.com/jmacd/essay/examples/internal/multishape"
	"github.com/jmacd/essay/num"
)

type (
	MRBuilder struct {
		points   []num.XYs
		title    string
		v        ms.Variable
		duration float64
		maxRate  float64
		colormap ms.Categories
		size     int
	}
)

func NewMultiRatePlot(points []num.XYs, v ms.Variable) MRBuilder {
	return MRBuilder{
		points: points,
		v:      v,
	}
}

func (r MRBuilder) Title(title string) MRBuilder {
	r.title = title
	return r
}

func (r MRBuilder) Duration(duration float64) MRBuilder {
	r.duration = duration
	return r
}

func (r MRBuilder) MaxRate(maxRate float64) MRBuilder {
	r.maxRate = maxRate
	return r
}

func (r MRBuilder) Size(size int) MRBuilder {
	r.size = size
	return r
}

func (r MRBuilder) ColorMap(colormap ms.Categories) MRBuilder {
	r.colormap = colormap
	return r
}

func (r MRBuilder) Build() num.Builder {
	pl := num.NewPlot().
		Size(r.size, r.size).
		Title(r.title).
		X(num.Axis().Range(0, r.duration).Hide()).
		Y(num.Axis().Range(0, r.maxRate).Hide())

	pal := HotToCoolDivergentPalette(len(r.colormap.Indexed))

	var lines []num.LBuilder

	for idx, points := range r.points {
		lines = append(lines, num.Line(points).
			Color(pal[idx]).
			Smooth().
			Width(lineWidth))
	}

	rand.Shuffle(len(lines), func(i, j int) {
		lines[i], lines[j] = lines[j], lines[i]
	})

	for _, line := range lines {
		pl = pl.Add(line)
	}

	return pl
}
