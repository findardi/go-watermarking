package watermark

import (
	"image"
	"image/color"
	"testing"
)

func TestGenerate(t *testing.T) {
	black := color.RGBA{0, 0, 0, 255}
	red := color.RGBA{255, 0, 0, 255}

	t.Run("watermark gambar di tengah, end-to-end", func(t *testing.T) {
		base := solid(100, 100, black)
		mark := ImageMark{Img: solid(40, 40, red), Scale: 0.2} // render → 20x20
		out, err := Generate(base, mark, Center{}, 1.0)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if out == nil {
			t.Fatal("hasil nil")
		}
		assertColor(t, out, 50, 50, red) // pusat: mark 20x20 di (40..59)
		assertColor(t, out, 0, 0, black) // pojok tak tersentuh
	})

	t.Run("error dari render mark diteruskan", func(t *testing.T) {
		base := solid(100, 100, black)
		mark := ImageMark{Img: solid(40, 40, red), Scale: 0} // render error
		if _, err := Generate(base, mark, Center{}, 1.0); err == nil {
			t.Error("ingin error dari render, dapat nil")
		}
	})

	t.Run("opacity invalid diteruskan", func(t *testing.T) {
		base := solid(100, 100, black)
		mark := ImageMark{Img: solid(40, 40, red), Scale: 0.2}
		if _, err := Generate(base, mark, Center{}, 1.5); err == nil {
			t.Error("ingin error opacity, dapat nil")
		}
	})
}

func TestPlacementRotation(t *testing.T) {
	if got := (Center{}).rotation(); got != 0 {
		t.Errorf("Center.rotation() = %d, want 0", got)
	}
	if got := (Pattern{Angle: 45}).rotation(); got != 45 {
		t.Errorf("Pattern.rotation() = %d, want 45", got)
	}
	if got := (Pattern{Angle: 45}).rotation(); got != 45 {
		t.Errorf("Pattern.rotation() = %d, want 45", got)
	}
}

func TestGenerateRotatesPattern(t *testing.T) {
	black := color.RGBA{0, 0, 0, 255}
	red := color.RGBA{255, 0, 0, 255}
	base := solid(100, 100, black)
	mark := ImageMark{Img: solid(40, 40, red), Scale: 0.4}

	straight, err := Generate(base, mark, Pattern{Angle: 0}, 1.0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	diagonal, err := Generate(base, mark, Pattern{Angle: 45}, 1.0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if imagesEqual(straight, diagonal) {
		t.Error("Pattern angle 45 = output identik dengan angle 0; rotasi tidak tersambung")
	}
}

func imagesEqual(a, b image.Image) bool {
	if a.Bounds() != b.Bounds() {
		return false
	}
	bd := a.Bounds()
	for y := bd.Min.Y; y < bd.Max.Y; y++ {
		for x := bd.Min.X; x < bd.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}
	return true
}
