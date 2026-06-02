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

func TestTextMarkRender(t *testing.T) {
	red := color.RGBA{255, 0, 0, 255}

	t.Run("menghasilkan kanvas berukuran wajar", func(t *testing.T) {
		m := TextMark{Text: "Hi", Scale: 0.1, Color: red}
		got, err := m.render(image.Pt(200, 200))
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		b := got.Bounds()
		if b.Dx() <= 0 || b.Dy() <= 0 {
			t.Fatalf("ukuran kanvas tidak wajar: %v", b)
		}
		// tinggi ~ Scale*baseW = 0.1*200 = 20px (toleran: metrik font + ceil)
		if b.Dy() < 15 || b.Dy() > 40 {
			t.Errorf("tinggi = %d, di luar kisaran wajar ~20", b.Dy())
		}
	})

	t.Run("teks lebih panjang = kanvas lebih lebar", func(t *testing.T) {
		short, _ := TextMark{Text: "I", Scale: 0.1, Color: red}.render(image.Pt(200, 200))
		long, _ := TextMark{Text: "IIIIIIII", Scale: 0.1, Color: red}.render(image.Pt(200, 200))
		if long.Bounds().Dx() <= short.Bounds().Dx() {
			t.Errorf("lebar teks panjang (%d) tidak > teks pendek (%d)",
				long.Bounds().Dx(), short.Bounds().Dx())
		}
	})

	t.Run("teks lebih panjang = kanvas lebih lebar", func(t *testing.T) {
		short, _ := TextMark{Text: "I", Scale: 0.1, Color: red}.render(image.Pt(200, 200))
		long, _ := TextMark{Text: "IIIIIIII", Scale: 0.1, Color: red}.render(image.Pt(200, 200))
		if long.Bounds().Dx() <= short.Bounds().Dx() {
			t.Errorf("lebar teks panjang (%d) tidak > teks pendek (%d)",
				long.Bounds().Dx(), short.Bounds().Dx())
		}
	})

	t.Run("ada piksel berwarna (teks benar tergambar)", func(t *testing.T) {
		m := TextMark{Text: "X", Scale: 0.2, Color: red}
		got, err := m.render(image.Pt(200, 200))
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if !hasOpaquePixel(got) {
			t.Error("tidak ada piksel buram; teks tidak tergambar?")
		}
	})

	t.Run("Scale<=0 atau teks kosong = error", func(t *testing.T) {
		if _, err := (TextMark{Text: "Hi", Scale: 0, Color: red}).render(image.Pt(200, 200)); err == nil {
			t.Error("Scale 0: ingin error")
		}
		if _, err := (TextMark{Text: "", Scale: 0.1, Color: red}).render(image.Pt(200, 200)); err == nil {
			t.Error("teks kosong: ingin error")
		}
	})

}

func hasOpaquePixel(img image.Image) bool {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a > 0 {
				return true
			}
		}
	}
	return false
}
