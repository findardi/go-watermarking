package app

import (
	"fmt"
	"go-watermarking/internal/codec"
	"go-watermarking/internal/watermark"
	"sync"
)

type MarkType string

const (
	MarkText  MarkType = "text"
	MarkImage MarkType = "image"
)

type Request struct {
	Image     [][]byte
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

type Result struct {
	Data   []byte `json:"data"`
	Format string `json:"format"`
}

func (s Service) Watermark(req Request) ([]Result, error) {
	mark, err := buildMark(req)
	if err != nil {
		return nil, err
	}

	placement, err := buildPlacement(req)
	if err != nil {
		return nil, err
	}

	result := make([]Result, len(req.Image))
	errs := make([]error, len(req.Image))

	const maxWorkers = 8
	buff := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for i, img := range req.Image {
		wg.Add(1)

		buff <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() {
				<-buff
			}()

			data, f, err := s.processOne(img, mark, placement, req.Opacity)
			if err != nil {
				errs[i] = err
				return
			}
			result[i] = Result{Data: data, Format: f}
		}()
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("failed process image %d: %w", i, err)
		}
	}

	return result, nil
}

func (s Service) processOne(image []byte, mark watermark.Mark, placement watermark.Placement, opacity float64) ([]byte, string, error) {
	base, format, err := codec.Decode(image)
	if err != nil {
		return nil, "", err
	}

	out, err := watermark.Generate(base, mark, placement, opacity)
	if err != nil {
		return nil, "", err
	}

	data, err := codec.Encode(out, format)
	if err != nil {
		return nil, "", err
	}

	return data, format, nil

}
