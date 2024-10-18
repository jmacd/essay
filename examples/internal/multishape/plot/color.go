package plot

import (
	"image/color"
	"sync"

	"github.com/lucasb-eyer/go-colorful"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/palette/brewer"
)

var (
	Black = color.RGBA{R: 0, G: 0, B: 0, A: 255}

	lock sync.Mutex
)

// http://mkweb.bcgsc.ca/brewer/swatches/brewer-palettes-swatches.pdf

func CoolToHotDivergentPalette(size int) []color.Color {
	return reverse(HotToCoolDivergentPalette(size))
}

func HotToCoolDivergentPalette(size int) []color.Color {
	var p pal
	p = pal{brewer.TypeDiverging, "PiYG"}
	p = pal{brewer.TypeDiverging, "BrBG"}
	p = pal{brewer.TypeDiverging, "PuOr"}
	p = pal{brewer.TypeDiverging, "PRGn"}
	return getOrInterpolate(p, size)
}

func CoolToHotSequentialPalette(size int) []color.Color {
	return reverse(HotToCoolSequentialPalette(size))
}

func HotToCoolSequentialPalette(size int) []color.Color {
	var p pal
	p = pal{brewer.TypeSequential, "PuBuGn"}
	p = pal{brewer.TypeDiverging, "PRGn"}
	p = pal{brewer.TypeDiverging, "PuOr"}
	p = pal{brewer.TypeDiverging, "RdBu"}
	p = pal{brewer.TypeDiverging, "RdYlBu"}
	p = pal{brewer.TypeSequential, "YlGnBu"}
	p = pal{brewer.TypeSequential, "YlGn"}
	p = pal{brewer.TypeSequential, "YlOrBr"}
	p = pal{brewer.TypeSequential, "Oranges"}
	p = pal{brewer.TypeDiverging, "RdYlBu"}
	return getOrInterpolate(p, size)
}

type pal struct {
	Type brewer.PaletteType
	Name string
}

func getOrInterpolate(p pal, number int) []color.Color {
	palN, err := brewer.GetPalette(p.Type, p.Name, number)
	if err == nil {
		return palN.Colors()
	}

	base := 11
	var palBase palette.Palette

	for base > 2 {
		palBase, err = brewer.GetPalette(p.Type, p.Name, base)
		if err != nil {
			base--
			continue
		}
		break
	}
	if err != nil {
		panic("Should have gotten at least 3 colors")
	}

	colBasein := palBase.Colors()
	colors := make([]color.Color, number)
	colors[0] = colBasein[0]
	colors[number-1] = colBasein[base-1]

	colBase := make([]colorful.Color, base)
	for i, in := range colBasein {
		colBase[i], _ = colorful.MakeColor(in)
	}

	unit := float64(base-1) / float64(number-1)

	for i := 1; i < number-1; i++ {
		value := float64(i) * unit
		baseBase := int(value)
		inter := value - float64(baseBase)

		colors[i] = colBase[baseBase].BlendLab(colBase[baseBase+1], inter/unit).Clamped()
	}

	return colors
}

func reverse(c []color.Color) []color.Color {
	for i := 0; i < len(c)/2; i++ {
		c[i], c[len(c)-1-i] = c[len(c)-1-i], c[i]
	}
	return c
}
