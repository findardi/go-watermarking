package watermark

import (
	"fmt"
	"image"
	"image/color"
	"math"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type Mark interface {
	render(baseSize image.Point) (image.Image, error)
}

type ImageMark struct {
	Img   image.Image
	Scale float64
}

func (m ImageMark) render(baseSize image.Point) (image.Image, error) {
	if m.Scale <= 0 {
		return nil, fmt.Errorf("scale must be > 0, got: %v", m.Scale)
	}

	srcW := m.Img.Bounds().Dx()
	srcH := m.Img.Bounds().Dy()

	tw := int(math.Round(m.Scale * float64(baseSize.X)))
	th := int(math.Round(float64(srcH) * float64(tw) / float64(srcW)))

	dst := image.NewRGBA(image.Rect(0, 0, tw, th))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), m.Img, m.Img.Bounds(), xdraw.Src, nil)

	return dst, nil
}

type TextMark struct {
	Text  string
	Scale float64
	Color color.Color
}

func (m TextMark) render(baseSize image.Point) (image.Image, error) {
	if m.Scale <= 0 {
		return nil, fmt.Errorf("scale must be > 0, got: %v", m.Scale)
	}

	if m.Text == "" {
		return nil, fmt.Errorf("text must not be empty")
	}

	px := m.Scale * float64(baseSize.X)

	ft, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(ft, &opentype.FaceOptions{
		Size: px,
		DPI:  72,
	})

	if err != nil {
		return nil, err
	}

	defer face.Close()

	d := &font.Drawer{Face: face}
	adv := d.MeasureString(m.Text)
	w := adv.Ceil()
	metrics := face.Metrics()
	h := (metrics.Ascent + metrics.Descent).Ceil()

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	d.Dst = dst
	d.Src = image.NewUniform(m.Color)
	d.Dot = fixed.P(0, metrics.Ascent.Ceil())
	d.DrawString(m.Text)

	return dst, nil
}
