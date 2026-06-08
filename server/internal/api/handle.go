package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-watermarking/internal/app"
	"io"
	"mime/multipart"
	"net/http"
)

const (
	maxRequestSize = 200 << 20 // 200mb
	maxImageSize   = 20 << 20  // 20mb
	maxFormMemory  = 32 << 20
)

type Watermarker interface {
	Watermark(context.Context, app.Request) ([]app.Result, []error)
}

type Handler struct {
	svc      Watermarker
	maxBytes int64
}

func NewHandler(svc Watermarker) *Handler {
	return &Handler{
		svc:      svc,
		maxBytes: maxRequestSize,
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

type imageResult struct {
	Index    int    `json:"index"`
	Filename string `json:"filename,omitempty"`
	Status   string `json:"status"`
	Format   string `json:"format,omitempty"`
	Data     []byte `json:"data,omitempty"`
	Error    string `json:"error,omitempty"`
}

type response struct {
	Total   int           `json:"total"`
	Success int           `json:"success"`
	Failed  int           `json:"failed"`
	Results []imageResult `json:"results"`
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
	defer func() {
		if err := r.MultipartForm.RemoveAll(); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}()

	var cfg Config
	if err := json.Unmarshal([]byte(r.FormValue("config")), &cfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid config json")
		return
	}

	img, names, err := readFilesPart(r, "image")
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

	result, errs := h.svc.Watermark(r.Context(), req)
	if result == nil {
		msg := "could not process request"
		if len(errs) > 0 && errs[0] != nil {
			msg = errs[0].Error()
		}
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}

	resp := response{
		Total:   len(result),
		Results: make([]imageResult, len(result)),
	}

	for i := range result {
		if i < len(errs) && errs[i] != nil {
			resp.Failed++
			resp.Results[i] = imageResult{
				Index:    i,
				Filename: names[i],
				Status:   "error",
				Error:    errs[i].Error(),
			}

			// todo create log
			continue
		}

		resp.Success++
		resp.Results[i] = imageResult{
			Index:    i,
			Filename: names[i],
			Status:   "ok",
			Format:   result[i].Format,
			Data:     result[i].Data,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.Failed > 0 && resp.Success > 0 {
		w.WriteHeader(http.StatusMultiStatus)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func readFilesPart(r *http.Request, name string) ([][]byte, []string, error) {
	headers := r.MultipartForm.File[name]

	if len(headers) == 0 {
		return nil, nil, errors.New("missing content-length")
	}
	files := make([][]byte, 0, len(headers))
	names := make([]string, 0, len(headers))
	for _, fh := range headers {
		if fh.Size > maxImageSize {
			return nil, nil, fmt.Errorf("image %q exceeds 20MB", fh.Filename)
		}
		data, err := readFile(fh)
		if err != nil {
			return nil, nil, err
		}
		files = append(files, data)
		names = append(names, fh.Filename)
	}

	return files, names, nil
}

func readFilePart(r *http.Request, name string) ([]byte, error) {
	headers := r.MultipartForm.File[name]

	if len(headers) == 0 {
		return nil, errors.New("missing content-length")
	}

	return readFile(headers[0])
}

func readFile(fh *multipart.FileHeader) (data []byte, err error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	return io.ReadAll(f)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	var b errorBody
	b.Error.Code = code
	b.Error.Message = msg

	_ = json.NewEncoder(w).Encode(b)
}
