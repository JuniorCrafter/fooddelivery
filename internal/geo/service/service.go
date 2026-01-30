package service

import (
	"context"
	"math"

	"github.com/JuniorCrafter/fooddelivery/internal/geo/repo"
)

type Service interface {
	UpdateLocation(ctx context.Context, courierID string, lat, lon float64) error
	GetDistance(ctx context.Context, courierID string, destLat, destLon float64) (float64, error)
}

type geoService struct {
	repo repo.GeoRepository
}

func New(r repo.GeoRepository) Service {
	return &geoService{repo: r}
}

func (s *geoService) UpdateLocation(ctx context.Context, courierID string, lat, lon float64) error {
	return s.repo.UpdateCourierLocation(ctx, courierID, lat, lon)
}

// GetDistance считает расстояние в километрах между курьером и точкой назначения
func (s *geoService) GetDistance(ctx context.Context, courierID string, destLat, destLon float64) (float64, error) {
	loc, err := s.repo.GetCourierLocation(ctx, courierID)
	if err != nil {
		return 0, err
	}

	// Математика Гаверсинуса
	const R = 6371 // Радиус Земли в км
	dLat := (destLat - loc.Latitude) * math.Pi / 180
	dLon := (destLon - loc.Longitude) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(loc.Latitude*math.Pi/180)*math.Cos(destLat*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c, nil
}
