package watermark

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

func Apply(base, mark image.Image, p Placement, opacity float64) (image.Image, error) {
	if opacity < 0 || opacity > 1 {
		return nil, fmt.Errorf("opacity %.2f out of bounds range [0,1]", opacity)
	}

	// copy original base
	out := image.NewRGBA(base.Bounds())
	draw.Draw(out, out.Bounds(), base, base.Bounds().Min, draw.Src)

	// mask
	mask := image.NewUniform(color.Alpha{A: uint8(opacity * 255)})
	markSize := mark.Bounds().Size()

	for _, pt := range p.positions(base.Bounds().Size(), markSize) {
		dst := image.Rect(pt.X, pt.Y, pt.X+markSize.X, pt.Y+markSize.Y).Add(base.Bounds().Min)
		draw.DrawMask(out, dst, mark, mark.Bounds().Min, mask, image.Point{}, draw.Over)
	}

	return out, nil
}
