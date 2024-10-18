package essay

import (
	"bytes"
	"image"
	"image/gif"

	"github.com/andybons/gogif"
	"github.com/jmacd/essay/internal/recovery"
)

const gifPeriods = 5

type (
	GBuilder struct {
		Images []EncodedImage
		Bounds image.Rectangle
	}
)

func Animation(images ...EncodedImage) GBuilder {
	g := GBuilder{}
	for _, img := range images {
		g = g.Add(img)
	}
	return g
}

func (g GBuilder) Add(i EncodedImage) GBuilder {
	g.Images = append(g.Images, i)
	g.Bounds = g.Bounds.Union(i.Bounds)
	return g
}

func (g GBuilder) Render(builtin Builtin) (interface{}, error) {
	defer recovery.Here()()
	outGif := &gif.GIF{}
	for _, simage := range g.Images {
		sbounds := simage.Bounds
		palettedImage := image.NewPaletted(sbounds, nil)
		img, err := simage.Decode()
		if err != nil {
			panic(err)
		}
		quantizer := gogif.MedianCutQuantizer{NumColor: 256}
		quantizer.Quantize(palettedImage, sbounds, img, image.ZP)

		outGif.Image = append(outGif.Image, palettedImage)
		outGif.Delay = append(outGif.Delay, 1)
	}
	outGif.Config.Width = g.Bounds.Dx()
	outGif.Config.Height = g.Bounds.Dy()
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, outGif); err != nil {
		panic(err)
	}
	ei := EncodedImage{
		Kind:   "gif",
		Bounds: g.Bounds,
		Data:   buf.Bytes(),
	}
	return builtin.RenderImage(ei)
}
