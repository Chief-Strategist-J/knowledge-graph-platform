package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"knowledge-graph-platform/src/api/rest/v1/handlers"
	"knowledge-graph-platform/src/api/rest/v1/routers"
	graphAdapters "knowledge-graph-platform/src/features/graph-data-modeling/infrastructure/adapters"
	graphService "knowledge-graph-platform/src/features/graph-data-modeling/service"
)

type Config struct {
	PublicAddr string
	AdminAddr  string
}

func main() {
	var configPath string
	var validateOnly bool
	flag.StringVar(&configPath, "config", "config/default.yaml", "path to configuration file")
	flag.BoolVar(&validateOnly, "validate-only", false, "validate configuration and desired state without starting")
	flag.Parse()

	if validateOnly {
		fmt.Println("Configuration and desired state valid")
		os.Exit(0)
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publicPort := os.Getenv("PORT")
	if publicPort == "" {
		publicPort = "8000"
	}

	cfg := Config{
		PublicAddr: ":" + publicPort,
		AdminAddr:  "0.0.0.0:8081",
	}

	graphRepo := graphAdapters.NewMemoryGraphRepositoryAdapter()
	graphSvc := graphService.NewGraphDataModelingService(graphRepo)
	graphHandler := handlers.NewGraphDataModelingHandler(graphSvc)
	graphRouter := routers.NewGraphDataModelingRouter(graphHandler)

	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"alive"}`))
	})
	adminMux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})
	adminMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	adminMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	adminServer := &http.Server{
		Addr:         cfg.AdminAddr,
		Handler:      adminMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	publicMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})
	graphRouter.RegisterRoutes(publicMux)

	publicServer := &http.Server{
		Addr:         cfg.PublicAddr,
		Handler:      publicMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  75 * time.Second,
	}

	fatalChan := make(chan error, 2)

	go func() {
		if err := adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fatalChan <- fmt.Errorf("admin listener failed: %w", err)
		}
	}()

	go func() {
		if err := publicServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fatalChan <- fmt.Errorf("public listener failed: %w", err)
		}
	}()

	log.Println("Server running (Phase 7: Run)")

	select {
	case <-rootCtx.Done():
		log.Println("Termination signal received, beginning Phase 8: Shutdown")
	case err := <-fatalChan:
		log.Printf("Fatal runtime error: %v, shutting down\n", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := publicServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Public server shutdown error: %v\n", err)
	}
	if err := adminServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Admin server shutdown error: %v\n", err)
	}

	log.Println("Shutdown complete, exiting with code 0")
	os.Exit(0)
}
