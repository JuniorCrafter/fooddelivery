package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/JuniorCrafter/fooddelivery/internal/courier/repo"
	"github.com/JuniorCrafter/fooddelivery/internal/courier/service"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/config"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/db"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/httpmw"
	"github.com/go-chi/chi/v5"
)

func main() {
	// 1. Загружаем конфиг (он сам поймет, где база: в Docker или локально)
	cfg := config.Load()

	// 2. Подключаемся к базе
	pool, err := db.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// 3. Инициализируем слои
	repository := repo.New(pool)
	courierServ := service.New(repository)

	r := chi.NewRouter()

	// Группа ручек, защищенных авторизацией
	r.Group(func(r chi.Router) {
		r.Use(httpmw.AuthMiddleware)

		// Посмотреть все свободные заказы
		r.Get("/courier/available", func(w http.ResponseWriter, r *http.Request) {
			orders, err := courierServ.FindWork(r.Context())
			if err != nil {
				http.Error(w, "Ошибка получения заказов", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(orders)
		})

		// Взять заказ в работу
		r.Post("/courier/status", func(w http.ResponseWriter, r *http.Request) {
			var input struct {
				OrderID int64  `json:"order_id"`
				Status  string `json:"status"` // Например: 'cooking', 'delivering', 'completed'
			}

			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "Неверный формат", http.StatusBadRequest)
				return
			}

			err := courierServ.ChangeStatus(r.Context(), input.OrderID, input.Status)
			if err != nil {
				http.Error(w, "Не удалось обновить статус", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "success",
				"message": "Статус заказа #" + strconv.FormatInt(input.OrderID, 10) + " изменен на " + input.Status,
			})
		})
		// Внутри r.Group(func(r chi.Router) {... })
		r.Get("/courier/dashboard/{id}", func(w http.ResponseWriter, r *http.Request) {
			idStr := chi.URLParam(r, "id")
			courierID, _ := strconv.ParseInt(idStr, 10, 64)

			summary, history, err := courierServ.GetDashboard(r.Context(), courierID)
			if err != nil {
				http.Error(w, "Ошибка получения данных", 500)
				return
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"summary": summary,
				"history": history,
			})
		})
	})

	// Запускаем сервис на порту 8083 (или берем из конфига)
	port := ":8083"
	log.Printf("Сервис курьеров успешно запущен на порту %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
}
