package main

import (
	"net/http"

	"food-delivery/internal/platform"
)

func main() {
	cfg := platform.LoadConfig("notifications")
	ready := platform.PostgresTCPReadyCheck(cfg)

	_ = platform.RunHTTP(cfg, func(mux *http.ServeMux) {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Notifications service is running\n"))
		})
	}, ready)
}
