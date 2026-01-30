package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient создает подключение к Redis
func NewRedisClient(addr string, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // если пароля нет, оставляем пустое ""
		DB:       0,        // используем стандартную базу №0
	})

	// Проверяем соединение (аналогично Ping в базе данных)
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
