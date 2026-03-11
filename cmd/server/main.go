package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wbhemingway/go-cartographer/internal/renderer"
)

func main() {
	cfg := renderer.DefaultConfig()
	engine := renderer.New(cfg)
	apiServer := NewServer(engine)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /render", apiServer.HandleRender)

	http.HandleFunc("POST /render", apiServer.HandleRender)

	port := ":8080"
	serve := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Starting map rendering server on http://localhost%s", port)
		if err := serve.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("\nShutdown signal received, draining ongoing requests...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := serve.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Cartographer Server stopped cleanly.")
}
