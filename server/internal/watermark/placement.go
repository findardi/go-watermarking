package watermark

import (
	"image"
)

const patternGap = 16

type Placement interface {
	positions(baseSize, markSize image.Point) []image.Point
}

type Center struct{}

func (Center) positions(baseSize, markSize image.Point) []image.Point {
	x := (baseSize.X - markSize.X) / 2
	y := (baseSize.Y - markSize.Y) / 2
	return []image.Point{image.Pt(x, y)}
}

type Pattern struct {
	Angle int
}

func (Pattern) positions(baseSize, markSize image.Point) []image.Point {
	stepX := markSize.X + patternGap
	stepY := markSize.Y + patternGap

	var points []image.Point
	for y := 0; y < baseSize.Y; y += stepY {
		for x := 0; x < baseSize.X; x += stepX {
			points = append(points, image.Pt(x, y))
		}
	}
	return points
}
