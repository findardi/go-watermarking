package app

  import (
        "bytes"
        "image"
        "image/color"
        "image/png"
        "testing"

        "go-watermarking/internal/codec"
  )

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

        t.Run("text mark, center — sukses & format dipertahankan", func(t *testing.T) {
                req := Request{
                        Image:     base,
                        MarkType:  MarkText,
                        Text:      "Hi",
                        HexColor:  "#FF0000",
                        Scale:     0.1,
                        Placement: "center",
                        Opacity:   1.0,
                }
                data, format, err := Service{}.Watermark(req)
                if err != nil {
                        t.Fatalf("err: %v", err)
                }
                if format != "png" {
                        t.Errorf("format = %q, want png", format)
                }
                // hasil harus gambar valid berukuran sama
                img, _, err := codec.Decode(data)
                if err != nil {
                        t.Fatalf("decode hasil: %v", err)
                }
                if img.Bounds() != image.Rect(0, 0, 100, 100) {
                        t.Errorf("bounds = %v, want 100x100", img.Bounds())
                }
        })

        t.Run("image mark, pattern — sukses", func(t *testing.T) {
                req := Request{
                        Image:     base,
                        MarkType:  MarkImage,
                        MarkImg:   pngBytes(t, 20, 20, color.RGBA{255, 0, 0, 255}),
                        Scale:     0.2,
                        Placement: "pattern",
                        Angle:     45,
                        Opacity:   0.5,
                }
                data, _, err := Service{}.Watermark(req)
                if err != nil {
                        t.Fatalf("err: %v", err)
                }
                if len(data) == 0 {
                        t.Fatal("hasil kosong")
                }
        })

        t.Run("base bukan gambar — error", func(t *testing.T) {
				req := Request{Image: []byte("bukan gambar"), MarkType: MarkText, Text: "x", HexColor: "#FFFFFF", Scale: 0.1, Placement: "center", Opacity: 1 }
                if _, _, err := (Service{}).Watermark(req); err == nil {
                        t.Error("ingin error")
                }
        })

        t.Run("mark type tak dikenal — error", func(t *testing.T) {
                req := Request{Image: base, MarkType: "video", Scale: 0.1, Placement: "center", Opacity: 1}
                if _, _, err := (Service{}).Watermark(req); err == nil {
                        t.Error("ingin error")
                }
        })

        t.Run("placement tak dikenal — error", func(t *testing.T) {
                req := Request{Image: base, MarkType: MarkText, Text: "x", HexColor: "#FFFFFF", Scale: 0.1, Placement: "diagonal", Opacity: 1}
                if _, _, err := (Service{}).Watermark(req); err == nil {
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
                {"FF0000", color.RGBA{}, true},   // tanpa #
                {"#FFF", color.RGBA{}, true},     // terlalu pendek
                {"#GGGGGG", color.RGBA{}, true},  // bukan hex
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