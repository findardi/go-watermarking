package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-watermarking/internal/app"
)

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("setup png: %v", err)
	}
	return buf.Bytes()
}

// buildMultipart merakit body multipart: config JSON + N image + (opsional) watermark.
func buildMultipart(t *testing.T, cfg string, imgs [][]byte, wm []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if cfg != "" {
		if err := mw.WriteField("config", cfg); err != nil {
			t.Fatal(err)
		}
	}
	for i, img := range imgs {
		fw, _ := mw.CreateFormFile("image", fmt.Sprintf("base%d.png", i))
		if _, err := fw.Write(img); err != nil {
			t.Fatal(err)
		}
	}
	if wm != nil {
		fw, _ := mw.CreateFormFile("watermark", "wm.png")
		if _, err := fw.Write(wm); err != nil {
			t.Fatal(err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	return &body, mw.FormDataContentType()
}

// stubService mengembalikan hasil/eror yang ditentukan, untuk menguji handler terisolasi.
type stubService struct {
	results []app.Result
	err     error
}

func (s stubService) Watermark(app.Request) ([]app.Result, error) {
	return s.results, s.err
}

func do(t *testing.T, h *Handler, body *bytes.Buffer, ct string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/watermark", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	h.Watermark(rec, req)
	return rec
}

func TestHandlerSuccess(t *testing.T) {
	// pakai service asli untuk happy path end-to-end
	h := NewHandler(app.Service{})
	cfg :=
		`{"mark":{"type":"text","text":"Hi","color":"#FF0000","scale":0.1},"placement":{"mode":"center"},"opacity":1.0}`
	body, ct := buildMultipart(t, cfg, [][]byte{pngBytes(t, 100, 100)}, nil)

	rec := do(t, h, body, ct)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("content-type = %q, want application/json", got)
	}

	var results []app.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("body bukan JSON valid: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Format != "png" {
		t.Errorf("format = %q, want png", results[0].Format)
	}
	// Data di-decode base64 otomatis oleh json.Unmarshal → harus gambar valid
	if _, _, err := image.Decode(bytes.NewReader(results[0].Data)); err != nil {
		t.Errorf("data bukan gambar valid: %v", err)
	}
}

func TestHandlerMultipleImages(t *testing.T) {
	h := NewHandler(app.Service{})
	cfg := `{"mark":{"type":"text","text":"Hi","color":"#FF0000","scale":0.1},"placement":{"mode":"center"},"opacity":1.0}`
	imgs := [][]byte{pngBytes(t, 80, 80), pngBytes(t, 40, 40)}
	body, ct := buildMultipart(t, cfg, imgs, nil)

	rec := do(t, h, body, ct)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var results []app.Result
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("body bukan JSON valid: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

func TestHandlerMissingImage(t *testing.T) {
	h := NewHandler(app.Service{})
	cfg := `{"mark":{"type":"text","text":"Hi","color":"#FF0000","scale":0.1},"placement":{"mode":"center"},"opacity":1.0}`
	body, ct := buildMultipart(t, cfg, nil, nil) // tanpa image
	rec := do(t, h, body, ct)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandlerInvalidConfig(t *testing.T) {
	h := NewHandler(app.Service{})
	body, ct := buildMultipart(t, `{bukan json`, [][]byte{pngBytes(t, 50, 50)}, nil)
	rec := do(t, h, body, ct)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandlerServiceError(t *testing.T) {
	h := NewHandler(stubService{err: io.ErrUnexpectedEOF})
	cfg := `{"mark":{"type":"text","text":"Hi","color":"#FF0000","scale":0.1},"placement":{"mode":"center"},"opacity":1.0}`
	body, ct := buildMultipart(t, cfg, [][]byte{pngBytes(t, 50, 50)}, nil)
	rec := do(t, h, body, ct)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rec.Code)
	}
}

func TestHandlerTooLarge(t *testing.T) {
	h := NewHandler(app.Service{})
	h.maxBytes = 100                                                          // paksa batas kecil
	body, ct := buildMultipart(t, `{}`, [][]byte{pngBytes(t, 200, 200)}, nil) // > 100 byte
	rec := do(t, h, body, ct)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", rec.Code)
	}
}
