package watermark

import (
	"image"
	"image/color"
	"testing"
)

func TestImageMarkRender(t *testing.T) {
	red := color.RGBA{255, 0, 0, 255}

	t.Run("scale 0.5 = ukuran benar & rasio terjaga", func(t *testing.T) {
		m := ImageMark{Img: solid(40, 20, red), Scale: 0.5}
		got, err := m.render(image.Pt(100, 100))
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		want := image.Rect(0, 0, 50, 25)
		if got.Bounds() != want {
			t.Errorf("bounds = %v, want %v", got.Bounds(), want)
		}
	})

	t.Run("warna solid tetap solid setelah scaling", func(t *testing.T) {
		m := ImageMark{Img: solid(40, 20, red), Scale: 0.5}
		got, err := m.render(image.Pt(100, 100))
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		assertColor(t, got, 25, 12, red)
	})

	t.Run("scale <= 0 = error", func(t *testing.T) {
		for _, s := range []float64{0, -0.3} {
			if _, err := (ImageMark{Img: solid(40, 20, red), Scale: s}).render(image.Pt(100, 100)); err == nil {
				t.Errorf("scale %v: ingin error, dapat nil", s)
			}
		}
	})
}
