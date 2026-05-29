package watermark

import (
	"image"
	"math"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

func rotate(src image.Image, angle int) image.Image {
	if angle == 0 {
		return src
	}

	b := src.Bounds()
	w, h := float64(b.Dx()), float64(b.Dy())
	rad := float64(angle) * math.Pi / 180
	s, c := math.Sin(rad), math.Cos(rad)

	newW := math.Abs((w * c) + (h * s))
	newH := math.Abs((w * s) + (h * c))

	cx, cy := w/2, h/2
	dx, dy := newW/2, newH/2

	m := f64.Aff3{
		c, -s, dx - (c*cx - s*cy),
		s, c, dy - (s*cx + c*cy),
	}

	dst := image.NewRGBA(image.Rect(0, 0, int(newW), int(newH)))
	xdraw.CatmullRom.Transform(dst, m, src, b, xdraw.Src, nil)

	return dst
}
