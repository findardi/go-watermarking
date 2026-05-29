package watermark

import (
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
