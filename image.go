package essay

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/jmacd/essay/internal/recovery"
)

const (
	PNG ImageKind = "png"
	SVG ImageKind = "svg"
)

type (
	ImageKind string

	EncodedImage struct {
		Kind   ImageKind
		Bounds image.Rectangle
		Data   []byte
	}
)

func (e Essay) RenderImage(img EncodedImage) (interface{}, error) {
	defer recovery.Here()()
	return e.execute("image.html", img)
}

func Image(i image.Image) EncodedImage {
	defer recovery.Here()()
	var buf bytes.Buffer

	if err := png.Encode(&buf, i); err != nil {
		panic(err)
	}
	return EncodedImage{
		Kind:   PNG,
		Bounds: i.Bounds(),
		Data:   buf.Bytes(),
	}
}

func (i EncodedImage) Render(builtin Builtin) (interface{}, error) {
	defer recovery.Here()()
	return builtin.RenderImage(i)
}

func (e EncodedImage) Decode() (image.Image, error) {
	switch e.Kind {
	case PNG:
		// This lets us take plot data from gonum and animate it,
		// have to re-parse the data though.
		return png.Decode(bytes.NewBuffer(e.Data))
	default:
		panic(fmt.Sprint("Unsupported decode: ", e.Kind))
	}
	return nil, nil
}
