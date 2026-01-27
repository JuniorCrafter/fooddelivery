package httpmw

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Секретный ключ должен быть таким же, как в сервисе Auth!
var secretKey = []byte("my_super_secret_key_123")

// AuthMiddleware — это и есть наш охранник
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Берем заголовок Authorization (там лежит наш "браслет")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
			return
		}

		// 2. Обычно заголовок выглядит так: "Bearer <токен>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Неверный формат заголовка", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// 3. Проверяем, настоящий ли токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Неверный или просроченный токен", http.StatusUnauthorized)
			return
		}

		// Если всё ок — пропускаем запрос дальше к "официанту"
		next.ServeHTTP(w, r)
	})
}
