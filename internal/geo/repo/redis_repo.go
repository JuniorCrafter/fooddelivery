package repo

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type GeoRepository interface {
	UpdateCourierLocation(ctx context.Context, courierID string, lat, lon float64) error
	// МЕНЯЕМ ТИП ТУТ: с GeoLocation на GeoPos
	GetCourierLocation(ctx context.Context, courierID string) (*redis.GeoPos, error)
}

type redisGeoRepo struct {
	client *redis.Client
}

func New(client *redis.Client) GeoRepository {
	return &redisGeoRepo{client: client}
}

func (r *redisGeoRepo) UpdateCourierLocation(ctx context.Context, courierID string, lat, lon float64) error {
	return r.client.GeoAdd(ctx, "couriers_locations", &redis.GeoLocation{
		Name:      courierID,
		Latitude:  lat,
		Longitude: lon,
	}).Err()
}

func (r *redisGeoRepo) GetCourierLocation(ctx context.Context, courierID string) (*redis.GeoPos, error) {
	// 1. Получаем результат. Он приходит как*redis.GeoPos
	res, err := r.client.GeoPos(ctx, "couriers_locations", courierID).Result()

	// 2. Проверяем ошибки и что список не пустой
	if err != nil {
		return nil, err
	}

	// Если курьер не найден, Redis вернет список, где первый элемент nil
	if len(res) == 0 || res == nil {
		return nil, fmt.Errorf("локация курьера %s не найдена", courierID)
	}

	// 3. Возвращаем именно ПЕРВЫЙ элемент списка
	return res[0], nil
}
