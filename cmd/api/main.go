package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi/v5"
)

// proxyHandler — это функция, которая перенаправляет запрос на другой адрес
func proxyHandler(targetAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target, _ := url.Parse(targetAddr)
		proxy := httputil.NewSingleHostReverseProxy(target)

		// Обновляем заголовки, чтобы микросервис понимал, откуда пришел запрос
		r.URL.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = target.Host

		proxy.ServeHTTP(w, r)
	}
}

func main() {
	r := chi.NewRouter()

	// Настраиваем маршруты: "путь" -> "адрес сервиса"
	// Мы используем http.StripPrefix, чтобы убрать "/auth" из пути перед отправкой
	// Пример: зашли на :8000/auth/login -> отправили на :8081/login
	r.Mount("/auth", http.StripPrefix("/auth", proxyHandler("http://localhost:8081")))
	r.Mount("/catalog", http.StripPrefix("/catalog", proxyHandler("http://localhost:8080")))
	r.Mount("/orders", http.StripPrefix("/orders", proxyHandler("http://localhost:8082")))
	r.Mount("/courier", http.StripPrefix("/courier", proxyHandler("http://localhost:8083")))
	r.Mount("/geo", http.StripPrefix("/geo", proxyHandler("http://localhost:8084")))

	log.Println("API Gateway запущен на порту :8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal(err)
	}
}
