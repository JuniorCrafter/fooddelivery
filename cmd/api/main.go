package main

import (
	"net/http"

	"github.com/JuniorCrafter/fooddelivery/internal/platform"
)

func main() {
	cfg := platform.LoadConfig("api")
	ready := platform.PostgresTCPReadyCheck(cfg)

	_ = platform.RunHTTP(cfg, func(mux *http.ServeMux) {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Food Delivery API service is running\n"))
		})
	}, ready)
}
