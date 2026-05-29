package watermark

import (
	"fmt"
	"image"
)

func Generate(base image.Image, m Mark, p Placement, opacity float64) (image.Image, error) {
	mark, err := m.render(base.Bounds().Size())
	if err != nil {
		return nil, fmt.Errorf("err render mark: %w", err)
	}
	return Apply(base, mark, p, opacity)
}
