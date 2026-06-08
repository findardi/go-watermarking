package app

import (
	"context"
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

type Service struct {
	sem chan struct{}
}

type Result struct {
	Data   []byte `json:"data"`
	Format string `json:"format"`
}

func NewService(maxWorkers int) Service {
	return Service{sem: make(chan struct{}, maxWorkers)}
}

func (s Service) Watermark(ctx context.Context, req Request) ([]Result, []error) {
	mark, err := buildMark(req)
	if err != nil {
		return nil, []error{err}
	}

	placement, err := buildPlacement(req)
	if err != nil {
		return nil, []error{err}
	}

	result := make([]Result, len(req.Image))
	errs := make([]error, len(req.Image))

	var wg sync.WaitGroup

	for i, img := range req.Image {
		select {
		case <-ctx.Done():
			errs[i] = ctx.Err()
			continue
		case s.sem <- struct{}{}:
		}

		wg.Add(1)

		go func(i int, img []byte) {
			defer wg.Done()
			defer func() {
				<-s.sem
			}()
			defer func() {
				if r := recover(); r != nil {
					errs[i] = fmt.Errorf("panic on image %d: %v", i, r)
				}
			}()

			data, f, err := s.processOne(img, mark, placement, req.Opacity)
			if err != nil {
				errs[i] = err
				return
			}
			result[i] = Result{Data: data, Format: f}
		}(i, img)
	}

	wg.Wait()
	return result, errs
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
