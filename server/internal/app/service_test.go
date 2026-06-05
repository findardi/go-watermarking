package app

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"

	"go-watermarking/internal/codec"
)

// firstErr mengembalikan error per-gambar pertama yang tidak nil.
// Service kini mengembalikan []error sejajar index, jadi slice non-nil
// berisi nil-nil bukan berarti gagal — yang menentukan adalah elemennya.
func firstErr(errs []error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

// pngBytes membuat gambar solid sebagai byte PNG untuk fixture.
func pngBytes(t *testing.T, w, h int, c color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("setup png: %v", err)
	}
	return buf.Bytes()
}

func TestServiceWatermark(t *testing.T) {
	black := color.RGBA{0, 0, 0, 255}
	base := pngBytes(t, 100, 100, black)

	t.Run("satu gambar, text mark, center — sukses & format dipertahankan", func(t *testing.T) {
		req := Request{
			Image:     [][]byte{base},
			MarkType:  MarkText,
			Text:      "Hi",
			HexColor:  "#FF0000",
			Scale:     0.1,
			Placement: "center",
			Opacity:   1.0,
		}
		results, errs := NewService(8).Watermark(context.Background(), req)
		if e := firstErr(errs); e != nil {
			t.Fatalf("err: %v", e)
		}
		if len(results) != 1 {
			t.Fatalf("len(results) = %d, want 1", len(results))
		}
		if results[0].Format != "png" {
			t.Errorf("format = %q, want png", results[0].Format)
		}
		// hasil harus gambar valid berukuran sama
		img, _, err := codec.Decode(results[0].Data)
		if err != nil {
			t.Fatalf("decode hasil: %v", err)
		}
		if img.Bounds() != image.Rect(0, 0, 100, 100) {
			t.Errorf("bounds = %v, want 100x100", img.Bounds())
		}
	})

	t.Run("banyak gambar — jumlah & urutan hasil terjaga", func(t *testing.T) {
		// ukuran berbeda agar urutan hasil bisa diverifikasi
		imgs := [][]byte{
			pngBytes(t, 100, 100, black),
			pngBytes(t, 60, 40, black),
			pngBytes(t, 30, 30, black),
		}
		req := Request{
			Image:     imgs,
			MarkType:  MarkText,
			Text:      "Hi",
			HexColor:  "#FF0000",
			Scale:     0.1,
			Placement: "center",
			Opacity:   1.0,
		}
		results, errs := NewService(8).Watermark(context.Background(), req)
		if e := firstErr(errs); e != nil {
			t.Fatalf("err: %v", e)
		}

		want := []image.Rectangle{
			image.Rect(0, 0, 100, 100),
			image.Rect(0, 0, 60, 40),
			image.Rect(0, 0, 30, 30),
		}
		if len(results) != len(want) {
			t.Fatalf("len(results) = %d, want %d", len(results), len(want))
		}
		for i, r := range results {
			img, _, err := codec.Decode(r.Data)
			if err != nil {
				t.Fatalf("results[%d] decode: %v", i, err)
			}
			if img.Bounds() != want[i] {
				t.Errorf("results[%d] bounds = %v, want %v", i, img.Bounds(), want[i])
			}
		}
	})

	t.Run("image mark, pattern — sukses", func(t *testing.T) {
		req := Request{
			Image:     [][]byte{base},
			MarkType:  MarkImage,
			MarkImg:   pngBytes(t, 20, 20, color.RGBA{255, 0, 0, 255}),
			Scale:     0.2,
			Placement: "pattern",
			Angle:     45,
			Opacity:   0.5,
		}
		results, errs := NewService(8).Watermark(context.Background(), req)
		if e := firstErr(errs); e != nil {
			t.Fatalf("err: %v", e)
		}
		if len(results) != 1 || len(results[0].Data) == 0 {
			t.Fatal("hasil kosong")
		}
	})

	t.Run("base bukan gambar — error", func(t *testing.T) {
		req := Request{Image: [][]byte{[]byte("bukan gambar")}, MarkType: MarkText, Text: "x", HexColor: "#FFFFFF", Scale: 0.1, Placement: "center", Opacity: 1}
		if _, errs := NewService(8).Watermark(context.Background(), req); firstErr(errs) == nil {
			t.Error("ingin error")
		}
	})

	t.Run("mark type tak dikenal — error", func(t *testing.T) {
		req := Request{Image: [][]byte{base}, MarkType: "video", Scale: 0.1, Placement: "center", Opacity: 1}
		if _, errs := NewService(8).Watermark(context.Background(), req); firstErr(errs) == nil {
			t.Error("ingin error")
		}
	})

	t.Run("placement tak dikenal — error", func(t *testing.T) {
		req := Request{Image: [][]byte{base}, MarkType: MarkText, Text: "x", HexColor: "#FFFFFF", Scale: 0.1, Placement: "diagonal", Opacity: 1}
		if _, errs := NewService(8).Watermark(context.Background(), req); firstErr(errs) == nil {
			t.Error("ingin error")
		}
	})
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		in      string
		want    color.RGBA
		wantErr bool
	}{
		{"#FF0000", color.RGBA{255, 0, 0, 255}, false},
		{"#00FF00", color.RGBA{0, 255, 0, 255}, false},
		{"#FFFFFF", color.RGBA{255, 255, 255, 255}, false},
		{"FF0000", color.RGBA{}, true},  // tanpa #
		{"#FFF", color.RGBA{}, true},    // terlalu pendek
		{"#GGGGGG", color.RGBA{}, true}, // bukan hex
		{"", color.RGBA{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseHexColor(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Errorf("%q: ingin error", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("%q: err tak terduga: %v", tt.in, err)
			}
			if got != color.Color(tt.want) {
				t.Errorf("%q = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
