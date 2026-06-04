package watermark

import (
	"image/color"
	"testing"
)

func TestRotate(t *testing.T) {
	red := color.RGBA{255, 0, 0, 255}

	t.Run("angle 0 = tidak berubah", func(t *testing.T) {
		src := solid(20, 20, red)
		got := rotate(src, 0)
		if got.Bounds() != src.Bounds() {
			t.Errorf("bounds = %v, want %v", got.Bounds(), src.Bounds())
		}
		assertColor(t, got, 10, 10, red)
	})

	t.Run("angle 45 = bbox membesar, pusat tetap merah, sudut transparan", func(t *testing.T) {
		src := solid(20, 20, red)
		got := rotate(src, 45)

		if got.Bounds().Dx() <= 20 {
			t.Errorf("Dx %d tidak membesar dari 20", got.Bounds().Dx())
		}

		b := got.Bounds()
		cx, cy := (b.Min.X+b.Max.X)/2, (b.Min.Y+b.Max.Y)/2
		assertColor(t, got, cx, cy, red) // pusat masih di dalam diamond → merah

		if a := at(t, got, b.Min.X, b.Min.Y).A; a != 0 {
			t.Errorf("sudut bbox alpha = %d, want 0 (transparan)", a)
		}
	})
}
