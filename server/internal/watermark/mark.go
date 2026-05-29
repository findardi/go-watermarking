package watermark

import (
	"fmt"
	"image"
	"math"

	xdraw "golang.org/x/image/draw"
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
