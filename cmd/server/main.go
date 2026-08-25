package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ashwanijha1405/url-shortener/internal/config"
	"github.com/Ashwanijha1405/url-shortener/internal/database"
	"github.com/Ashwanijha1405/url-shortener/internal/handler"
	"github.com/Ashwanijha1405/url-shortener/internal/repository/postgres"
)

func main() {
	ctx := context.Background()

	// Connect to PostgreSQL
	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Pool.Close()

	log.Println("connected to PostgreSQL")

	// Create repository
	repo := postgres.NewRepository(db)

	// Inject repository into handler
	h := handler.NewHandler(repo)

	cfg := config.Load()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/api/v1/urls", h.CreateURL)
	mux.HandleFunc("/{shortCode}", h.RedirectURL)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// Start server in a separate goroutine.
	go func() {
		log.Printf("starting server on :%s", cfg.Port)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for termination signal.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan

	log.Println("shutdown signal received")

	// Give active requests some time to finish.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		return
	}

	log.Println("server stopped gracefully")
}
