package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Наш секретный ключ (в реальном проекте он должен быть в.env)
var secretKey = []byte("my_super_secret_key_123")

// GenerateToken создает зашифрованную строку с ID пользователя
func GenerateToken(userID int64, role string) (string, error) {
	// Создаем "полезную нагрузку" (claims)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,                                  // Роль: client, courier или admin
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Токен живет 24 часа
	})

	// Подписываем токен нашим ключом
	return token.SignedString(secretKey)
}
