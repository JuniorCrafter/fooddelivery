package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JuniorCrafter/fooddelivery/internal/geo/repo"
	"github.com/JuniorCrafter/fooddelivery/internal/geo/service"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/cache"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/config"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()

	// 2. Берем адрес Redis из настроек (в Docker это будет "redis:6379")
	rdb, err := cache.NewRedisClient(cfg.RedisAddr, "")
	if err != nil {
		log.Fatalf("Ошибка Redis: %v", err)
	}

	repository := repo.New(rdb)
	geoServ := service.New(repository)

	r := chi.NewRouter()

	// 1. Ручка для курьера: отправка координат
	r.Post("/geo/update", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			CourierID string  `json:"courier_id"`
			Lat       float64 `json:"lat"`
			Lon       float64 `json:"lon"`
		}
		json.NewDecoder(r.Body).Decode(&input)
		geoServ.UpdateLocation(r.Context(), input.CourierID, input.Lat, input.Lon)
		w.Write([]byte("Координаты обновлены"))
	})

	// 2. Ручка для расчета дистанции (например, до ресторана)
	r.Get("/geo/distance", func(w http.ResponseWriter, r *http.Request) {
		courierID := r.URL.Query().Get("courier_id")
		// Пример координат ресторана: 55.75, 37.61 (Москва)
		dist, err := geoServ.GetDistance(r.Context(), courierID, 55.75, 37.61)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]float64{"distance_km": dist})
	})

	log.Println("Geo Service запущен на порту :8084")
	http.ListenAndServe(":8084", r)
}
