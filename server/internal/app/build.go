package app

import (
	"fmt"
	"go-watermarking/internal/codec"
	"go-watermarking/internal/watermark"
	"image/color"
)

func buildMark(req Request) (watermark.Mark, error) {
	switch req.MarkType {
	case MarkText:
		c, err := parseHexColor(req.HexColor)
		if err != nil {
			return nil, err
		}
		return watermark.TextMark{Text: req.Text, Scale: req.Scale, Color: c}, nil

	case MarkImage:
		img, _, err := codec.Decode(req.MarkImg)
		if err != nil {
			return nil, err
		}
		return watermark.ImageMark{Img: img, Scale: req.Scale}, nil

	default:
		return nil, fmt.Errorf("unknown mark type: %q", req.MarkType)
	}
}

func buildPlacement(req Request) (watermark.Placement, error) {
	switch req.Placement {
	case "center":
		return watermark.Center{}, nil
	case "pattern":
		return watermark.Pattern{Angle: req.Angle}, nil
	default:
		return nil, fmt.Errorf("unknown placement: %q", req.Placement)
	}
}

func parseHexColor(s string) (color.Color, error) {
	var r, g, b uint8

	n, err := fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	if err != nil || n != 3 {
		return nil, fmt.Errorf("invalid color hex: %q", s)
	}

	return color.RGBA{r, g, b, 255}, nil
}
