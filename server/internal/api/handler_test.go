package api

import (
	"bytes"
	"context"
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
// results & errs sejajar index, meniru kontrak Service.Watermark.
type stubService struct {
	results []app.Result
	errs    []error
}

func (s stubService) Watermark(context.Context, app.Request) ([]app.Result, []error) {
	return s.results, s.errs
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
	h := NewHandler(app.NewService(8))
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

	var resp response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body bukan JSON valid: %v", err)
	}
	if resp.Total != 1 || resp.Success != 1 || resp.Failed != 0 {
		t.Fatalf("envelope = %+v, want total=1 success=1 failed=0", resp)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(resp.Results))
	}
	got := resp.Results[0]
	if got.Status != "ok" {
		t.Errorf("status = %q, want ok", got.Status)
	}
	if got.Format != "png" {
		t.Errorf("format = %q, want png", got.Format)
	}
	// Data di-decode base64 otomatis oleh json.Unmarshal → harus gambar valid
	if _, _, err := image.Decode(bytes.NewReader(got.Data)); err != nil {
		t.Errorf("data bukan gambar valid: %v", err)
	}
}

func TestHandlerMultipleImages(t *testing.T) {
	h := NewHandler(app.NewService(8))
	cfg := `{"mark":{"type":"text","text":"Hi","color":"#FF0000","scale":0.1},"placement":{"mode":"center"},"opacity":1.0}`
	imgs := [][]byte{pngBytes(t, 80, 80), pngBytes(t, 40, 40)}
	body, ct := buildMultipart(t, cfg, imgs, nil)

	rec := do(t, h, body, ct)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var resp response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body bukan JSON valid: %v", err)
	}
	if resp.Total != 2 || resp.Success != 2 || resp.Failed != 0 {
		t.Fatalf("envelope = %+v, want total=2 success=2 failed=0", resp)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(resp.Results))
	}
}

func TestHandlerMissingImage(t *testing.T) {
	h := NewHandler(app.NewService(8))
	cfg := `{"mark":{"type":"text","text":"Hi","color":"#FF0000","scale":0.1},"placement":{"mode":"center"},"opacity":1.0}`
	body, ct := buildMultipart(t, cfg, nil, nil) // tanpa image
	rec := do(t, h, body, ct)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandlerInvalidConfig(t *testing.T) {
	h := NewHandler(app.NewService(8))
	body, ct := buildMultipart(t, `{bukan json`, [][]byte{pngBytes(t, 50, 50)}, nil)
	rec := do(t, h, body, ct)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// Kegagalan tingkat-request: service mengembalikan results==nil → 422.
func TestHandlerServiceError(t *testing.T) {
	h := NewHandler(stubService{errs: []error{io.ErrUnexpectedEOF}})
	cfg := `{"mark":{"type":"text","text":"Hi","color":"#FF0000","scale":0.1},"placement":{"mode":"center"},"opacity":1.0}`
	body, ct := buildMultipart(t, cfg, [][]byte{pngBytes(t, 50, 50)}, nil)
	rec := do(t, h, body, ct)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rec.Code)
	}
}

// Partial success (inti Opsi A): sebagian gambar sukses, sebagian gagal →
// 207 Multi-Status, tiap item membawa status/error-nya sendiri.
func TestHandlerPartialSuccess(t *testing.T) {
	okImg := pngBytes(t, 30, 30)
	h := NewHandler(stubService{
		results: []app.Result{
			{Data: okImg, Format: "png"},
			{}, // index gagal: data kosong
		},
		errs: []error{nil, io.ErrUnexpectedEOF},
	})
	cfg := `{"mark":{"type":"text","text":"Hi","color":"#FF0000","scale":0.1},"placement":{"mode":"center"},"opacity":1.0}`
	// dua part image → handler punya dua filename sejajar index hasil
	body, ct := buildMultipart(t, cfg, [][]byte{okImg, pngBytes(t, 20, 20)}, nil)

	rec := do(t, h, body, ct)

	if rec.Code != http.StatusMultiStatus {
		t.Fatalf("status = %d, want 207; body=%s", rec.Code, rec.Body.String())
	}

	var resp response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body bukan JSON valid: %v", err)
	}
	if resp.Total != 2 || resp.Success != 1 || resp.Failed != 1 {
		t.Fatalf("envelope = %+v, want total=2 success=1 failed=1", resp)
	}
	if resp.Results[0].Status != "ok" || resp.Results[0].Format != "png" {
		t.Errorf("results[0] = %+v, want status=ok format=png", resp.Results[0])
	}
	if resp.Results[1].Status != "error" || resp.Results[1].Error == "" {
		t.Errorf("results[1] = %+v, want status=error dengan pesan", resp.Results[1])
	}
}

func TestHandlerTooLarge(t *testing.T) {
	h := NewHandler(app.NewService(8))
	h.maxBytes = 100                                                          // paksa batas kecil
	body, ct := buildMultipart(t, `{}`, [][]byte{pngBytes(t, 200, 200)}, nil) // > 100 byte
	rec := do(t, h, body, ct)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", rec.Code)
	}
}
