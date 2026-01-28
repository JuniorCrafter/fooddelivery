package service

import (
	"context"

	"github.com/JuniorCrafter/fooddelivery/internal/courier/repo"
)

type Service interface {
	FindWork(ctx context.Context) ([]repo.OrderInfo, error)
	TakeOrder(ctx context.Context, courierID, orderID int64) error
}

type courierService struct {
	repo repo.Repository
}

func New(r repo.Repository) Service {
	return &courierService{repo: r}
}

func (s *courierService) FindWork(ctx context.Context) ([]repo.OrderInfo, error) {
	return s.repo.GetNewOrders(ctx)
}

func (s *courierService) TakeOrder(ctx context.Context, courierID, orderID int64) error {
	return s.repo.AcceptOrder(ctx, courierID, orderID)
}
