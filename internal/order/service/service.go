package service

import (
	"context"

	"github.com/JuniorCrafter/fooddelivery/internal/order/repo"
)

type Service interface {
	PlaceOrder(ctx context.Context, userID int64, items []repo.OrderItem) (int64, error)
}

type orderService struct {
	repo repo.Repository
}

func New(r repo.Repository) Service {
	return &orderService{repo: r}
}

func (s *orderService) PlaceOrder(ctx context.Context, userID int64, items []repo.OrderItem) (int64, error) {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	order := repo.Order{
		UserID:     userID,
		TotalPrice: total,
		Items:      items,
	}

	return s.repo.CreateOrder(ctx, order)
}
