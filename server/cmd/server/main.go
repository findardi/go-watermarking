package main

import (
	"context"
	"errors"
	"go-watermarking/internal/api"
	"go-watermarking/internal/app"
	"go-watermarking/internal/web"
	"log"
	"net/http"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const devOrigin = "http://localhost:5173"

func main() {
	service := app.NewService(runtime.NumCPU())
	handler := api.NewHandler(service)

	mux := http.NewServeMux()
	webFs, err := web.FS()
	if err != nil {
		log.Fatalf("web fs: %v", err)
	}

	mux.Handle("/", api.SPA(webFs))
	mux.HandleFunc("POST /api/v1/watermark", handler.Watermark)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           api.CORS(devOrigin, mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer func() { stop() }()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() { cancel() }()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shhutdown: %v", err)
	}

	log.Println("server stoped")
}
