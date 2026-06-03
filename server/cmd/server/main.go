package main

import (
	"context"
	"errors"
	"go-watermarking/internal/api"
	"go-watermarking/internal/app"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	handler := api.NewHandler(app.Service{})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/watermark", handler.Watermark)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
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
