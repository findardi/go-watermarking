package codec

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

// makeImage membuat gambar solid kecil untuk fixture.
func makeImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{200, 100, 50, 255})
		}
	}
	return img
}

func encodePNG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, makeImage()); err != nil {
		t.Fatalf("setup png: %v", err)
	}
	return buf.Bytes()
}

func encodeJPEG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, makeImage(), nil); err != nil {
		t.Fatalf("setup jpeg: %v", err)
	}
	return buf.Bytes()
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantFormat string
	}{
		{"png", encodePNG(t), "png"},
		{"jpeg", encodeJPEG(t), "jpeg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, format, err := Decode(tt.data)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if format != tt.wantFormat {
				t.Errorf("format = %q, want %q", format, tt.wantFormat)
			}
			if img.Bounds() != image.Rect(0, 0, 8, 8) {
				t.Errorf("bounds = %v, want 8x8", img.Bounds())
			}
		})
	}
}

func TestDecodeInvalid(t *testing.T) {
	if _, _, err := Decode([]byte("ini bukan gambar")); err == nil {
		t.Error("ingin error untuk data bukan gambar, dapat nil")
	}
}

func TestEncode(t *testing.T) {
	for _, format := range []string{"png", "jpeg"} {
		t.Run(format, func(t *testing.T) {
			data, err := Encode(makeImage(), format)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("hasil encode kosong")
			}
			// round-trip: hasil encode harus bisa di-decode balik ke format sama
			_, gotFormat, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("decode balik: %v", err)
			}
			if gotFormat != format {
				t.Errorf("round-trip format = %q, want %q", gotFormat, format)
			}
		})
	}
}

func TestEncodeUnsupported(t *testing.T) {
	if _, err := Encode(makeImage(), "gif"); err == nil {
		t.Error("ingin error untuk format tak didukung, dapat nil")
	}
}
