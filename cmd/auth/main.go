package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JuniorCrafter/fooddelivery/internal/auth/handler"
	"github.com/JuniorCrafter/fooddelivery/internal/auth/repo"
	"github.com/JuniorCrafter/fooddelivery/internal/auth/service"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/config"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	r := repo.New(pool)
	svc := service.New(r, []byte(cfg.JWTSecret), cfg.AccessTTL, cfg.RefreshTTL, cfg.BcryptCost)
	h := handler.New(svc, []byte(cfg.JWTSecret))

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("auth listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	_ = srv.Shutdown(shutdownCtx)
	log.Println("auth stopped")
}
