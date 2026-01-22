package platform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ReadyCheck func(ctx context.Context) error

type HealthPayload struct {
	Status  string    `json:"status"`
	Service string    `json:"service"`
	Time    time.Time `json:"time"`
}

func RunHTTP(cfg Config, register func(mux *http.ServeMux), ready ReadyCheck) error {
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	mux := http.NewServeMux()

	// Built-in endpoints for container orchestration and manual checks.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HealthPayload{Status: "ok", Service: cfg.ServiceName, Time: time.Now().UTC()})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if ready == nil {
			writeJSON(w, http.StatusOK, HealthPayload{Status: "ready", Service: cfg.ServiceName, Time: time.Now().UTC()})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := ready(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status":  "not_ready",
				"service": cfg.ServiceName,
				"time":    time.Now().UTC(),
				"error":   err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, HealthPayload{Status: "ready", Service: cfg.ServiceName, Time: time.Now().UTC()})
	})

	// Custom service routes.
	if register != nil {
		register(mux)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           loggingMiddleware(cfg, mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("[%s] http listening on %s", cfg.ServiceName, addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Printf("[%s] shutdown signal received", cfg.ServiceName)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Printf("[%s] stopped", cfg.ServiceName)
	return nil
}

func PostgresTCPReadyCheck(cfg Config) ReadyCheck {
	if !cfg.CheckPostgres {
		return nil
	}
	return func(ctx context.Context) error {
		d := net.Dialer{Timeout: 1 * time.Second}
		conn, err := d.DialContext(ctx, "tcp", cfg.PostgresAddr)
		if err != nil {
			return fmt.Errorf("postgres not reachable (%s): %w", cfg.PostgresAddr, err)
		}
		_ = conn.Close()
		return nil
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func loggingMiddleware(cfg Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s ua=%q dur=%s", cfg.ServiceName, r.Method, r.URL.Path, r.UserAgent(), time.Since(start))
	})
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.LUTC)
}
