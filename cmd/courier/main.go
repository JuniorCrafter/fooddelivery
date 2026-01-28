package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/JuniorCrafter/fooddelivery/internal/courier/repo"
	"github.com/JuniorCrafter/fooddelivery/internal/courier/service"
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
	courierServ := service.New(repository)

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(httpmw.AuthMiddleware)

		// Посмотреть доступные заказы
		r.Get("/courier/available", func(w http.ResponseWriter, r *http.Request) {
			orders, _ := courierServ.FindWork(r.Context())
			json.NewEncoder(w).Encode(orders)
		})

		// Взять заказ в работу
		r.Post("/courier/accept", func(w http.ResponseWriter, r *http.Request) {
			var input struct {
				CourierID int64 `json:"courier_id"`
				OrderID   int64 `json:"order_id"`
			}
			json.NewDecoder(r.Body).Decode(&input)

			err := courierServ.TakeOrder(r.Context(), input.CourierID, input.OrderID)
			if err != nil {
				http.Error(w, "Не удалось взять заказ", 400)
				return
			}
			w.Write([]byte("Заказ принят!"))
		})
	})

	log.Println("Сервис курьеров запущен на порту :8083")
	http.ListenAndServe(":8083", r)
}
