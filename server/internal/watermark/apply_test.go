package watermark

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func solid(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
	return img
}

func at(t *testing.T, img image.Image, x, y int) color.RGBA {
	t.Helper()
	r, g, b, a := img.At(x, y).RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func assertColor(t *testing.T, img image.Image, x, y int, want color.RGBA) {
	t.Helper()
	if got := at(t, img, x, y); got != want {
		t.Errorf("piksel (%d,%d) = %v, want %v", x, y, got, want)
	}
}

func near(a, b uint8, tol int) bool {
	d := int(a) - int(b)
	if d < 0 {
		d = -d
	}
	return d <= tol
}

func TestApply(t *testing.T) {
	black := color.RGBA{0, 0, 0, 255}
	red := color.RGBA{255, 0, 0, 255}

	t.Run("opacity penuh menimpa di tengah", func(t *testing.T) {
		out, err := Apply(solid(10, 10, black), solid(4, 4, red), Center{}, 1.0)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if out == nil {
			t.Fatal("hasil nil")
		}
		assertColor(t, out, 3, 3, red)
		assertColor(t, out, 6, 6, red)
		assertColor(t, out, 0, 0, black)
	})

	t.Run("opacity 0 = mark tak terlihat", func(t *testing.T) {
		out, err := Apply(solid(10, 10, black), solid(4, 4, red), Center{}, 0)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		assertColor(t, out, 5, 5, black)
	})

	t.Run("opacity 0.5 = warna berbaur ~ setengah", func(t *testing.T) {
		out, err := Apply(solid(10, 10, black), solid(4, 4, red), Center{}, 0.5)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		c := at(t, out, 4, 4)
		if !near(c.R, 128, 2) || c.G != 0 || c.B != 0 {
			t.Errorf("warna tengah = %v, ingin R~128 G0 B0", c)
		}
	})

	t.Run("base tidak dimutasi", func(t *testing.T) {
		base := solid(10, 10, black)
		if _, err := Apply(base, solid(4, 4, red), Center{}, 1.0); err != nil {
			t.Fatalf("err: %v", err)
		}
		assertColor(t, base, 5, 5, black)
	})

	t.Run("pattern menempel di tiap titik grid", func(t *testing.T) {
		out, err := Apply(solid(50, 50, black), solid(10, 10, red), Pattern{}, 1.0)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		assertColor(t, out, 0, 0, red)
		assertColor(t, out, 26, 26, red)
		assertColor(t, out, 20, 20, black)
	})

	t.Run("opacity di luar [0,1] = error", func(t *testing.T) {
		for _, op := range []float64{-0.1, 1.5} {
			if _, err := Apply(solid(10, 10, black), solid(4, 4, red), Center{}, op); err == nil {
				t.Errorf("opacity %v: ingin error, dapat nil", op)
			}
		}
	})

}
