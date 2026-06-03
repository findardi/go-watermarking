package app

import (
	"go-watermarking/internal/codec"
	"go-watermarking/internal/watermark"
)

type MarkType string

const (
	MarkText  MarkType = "text"
	MarkImage MarkType = "image"
)

type Request struct {
	Image     []byte
	MarkType  MarkType
	Text      string
	HexColor  string
	MarkImg   []byte
	Scale     float64
	Placement string
	Angle     int
	Opacity   float64
}

type Service struct{}

func (s Service) Watermark(req Request) (data []byte, format string, err error) {
	base, format, err := codec.Decode(req.Image)
	if err != nil {
		return nil, "", err
	}

	mark, err := buildMark(req)
	if err != nil {
		return nil, "", err
	}

	placement, err := buildPlacement(req)
	if err != nil {
		return nil, "", err
	}

	out, err := watermark.Generate(base, mark, placement, req.Opacity)
	if err != nil {
		return nil, "", err
	}

	data, err = codec.Encode(out, format)
	if err != nil {
		return nil, "", err
	}

	return data, format, nil
}
