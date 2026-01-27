package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/JuniorCrafter/fooddelivery/internal/catalog/repo/pg"
	"github.com/JuniorCrafter/fooddelivery/internal/catalog/service"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/config"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/db"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/httpmw"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()

	// Подключаемся к базе (используем тот же URL, что и в Auth)
	pool, err := db.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	repository := pg.New(pool)
	catService := service.New(repository)

	r := chi.NewRouter()

	// Открытая ручка: список товаров могут смотреть все
	r.Get("/products", func(w http.ResponseWriter, r *http.Request) {
		products, _ := catService.GetAllProducts(r.Context())
		json.NewEncoder(w).Encode(products)
	})

	// ЗАЩИЩЕННАЯ группа: только для тех, у кого есть токен
	r.Group(func(r chi.Router) {
		r.Use(httpmw.AuthMiddleware) // Вешаем нашего охранника на эту группу

		r.Post("/products", func(w http.ResponseWriter, r *http.Request) {
			var p pg.Product
			json.NewDecoder(r.Body).Decode(&p)
			id, err := catService.AddProduct(r.Context(), p)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			json.NewEncoder(w).Encode(map[string]int64{"id": id})
		})
	})

	log.Println("Сервис каталога запущен на порту :8080")
	http.ListenAndServe(":8080", r)
}
