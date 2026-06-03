package api

import (
	"encoding/json"
	"errors"
	"go-watermarking/internal/app"
	"io"
	"net/http"
)

const defaultMaxBytes = 10 << 20

type Watermarker interface {
	Watermark(app.Request) (data []byte, format string, err error)
}

type Handler struct {
	svc      Watermarker
	maxBytes int64
}

func NewHandler(svc Watermarker) *Handler {
	return &Handler{
		svc:      svc,
		maxBytes: defaultMaxBytes,
	}
}

type Config struct {
	Mark      Mark      `json:"mark"`
	Placement Placement `json:"placement"`
	Opacity   float64   `json:"opacity"`
}

type Mark struct {
	Type  string  `json:"type"`
	Text  string  `json:"text"`
	Color string  `json:"color"`
	Scale float64 `json:"scale"`
}

type Placement struct {
	Mode  string `json:"mode"`
	Angle int    `json:"angle"`
}

type errorBody struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (h *Handler) Watermark(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)

	err := r.ParseMultipartForm(h.maxBytes)

	var mbe *http.MaxBytesError
	if errors.As(err, &mbe) {
		writeError(w, http.StatusRequestEntityTooLarge, "request too large")
		return
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	var cfg Config
	if err := json.Unmarshal([]byte(r.FormValue("config")), &cfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid config json")
		return
	}

	img, err := readFilePart(r, "image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing image")
		return
	}

	var markImg []byte
	if cfg.Mark.Type == string(app.MarkImage) {
		markImg, err = readFilePart(r, "watermark")
		if err != nil {
			writeError(w, http.StatusBadRequest, "missing watermark image")
			return
		}
	}

	req := app.Request{
		Image:     img,
		MarkType:  app.MarkType(cfg.Mark.Type),
		Text:      cfg.Mark.Text,
		HexColor:  cfg.Mark.Color,
		MarkImg:   markImg,
		Scale:     cfg.Mark.Scale,
		Placement: cfg.Placement.Mode,
		Angle:     cfg.Placement.Angle,
		Opacity:   cfg.Opacity,
	}

	data, format, err := h.svc.Watermark(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	w.Header().Set("Content-Type", contentType(format))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func readFilePart(r *http.Request, name string) ([]byte, error) {
	f, _, err := r.FormFile(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return io.ReadAll(f)
}

func contentType(format string) string {
	switch format {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	var b errorBody
	b.Error.Code = code
	b.Error.Message = msg

	_ = json.NewEncoder(w).Encode(b)
}
