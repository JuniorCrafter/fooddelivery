package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	// Создадим позже или напишем тут
	"github.com/JuniorCrafter/fooddelivery/internal/order/repo"
	"github.com/JuniorCrafter/fooddelivery/internal/order/service"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/config"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/db"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/httpmw"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()
	pool, err := db.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	repository := repo.New(pool)
	orderService := service.New(repository)

	r := chi.NewRouter()

	// Защищаем ручку создания заказа
	r.Group(func(r chi.Router) {
		r.Use(httpmw.AuthMiddleware)

		r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
			var input struct {
				UserID int64            `json:"user_id"`
				Items  []repo.OrderItem `json:"items"`
			}
			json.NewDecoder(r.Body).Decode(&input)

			id, err := orderService.PlaceOrder(r.Context(), input.UserID, input.Items)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]int64{"order_id": id})
		})
	})

	log.Println("Сервис заказов запущен на порту :8082")
	http.ListenAndServe(":8082", r)
}
