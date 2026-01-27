package main

import (
	"context"
	"log"
	"net/http"

	"github.com/JuniorCrafter/fooddelivery/internal/auth/handler"
	"github.com/JuniorCrafter/fooddelivery/internal/auth/repo"
	"github.com/JuniorCrafter/fooddelivery/internal/auth/service"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/config"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/db"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()

	log.Printf("Сервис авторизации пробует подключиться к: %s", cfg.DatabaseURL)

	// Подключаемся к базе
	pool, err := db.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Критическая ошибка БД: %v", err)
	}

	// Собираем микросервис
	repository := repo.New(pool)
	authService := service.New(repository)
	authHandler := handler.New(authService)

	r := chi.NewRouter()
	authHandler.RegisterRoutes(r)

	log.Printf("Сервис успешно запущен на порту %s", cfg.AuthPort)
	if err := http.ListenAndServe(cfg.AuthPort, r); err != nil {
		log.Fatal(err)
	}
}
