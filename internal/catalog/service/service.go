package service

import (
	"context"
	"errors"

	"github.com/JuniorCrafter/fooddelivery/internal/catalog/repo/pg"
)

type Service interface {
	AddProduct(ctx context.Context, p pg.Product) (int64, error)
	GetAllProducts(ctx context.Context) ([]pg.Product, error)
}

type catalogService struct {
	repo pg.Repository
}

func New(r pg.Repository) Service {
	return &catalogService{repo: r}
}

func (s *catalogService) AddProduct(ctx context.Context, p pg.Product) (int64, error) {
	if p.Price <= 0 {
		return 0, errors.New("цена должна быть больше нуля")
	}
	return s.repo.Create(ctx, p)
}

func (s *catalogService) GetAllProducts(ctx context.Context) ([]pg.Product, error) {
	return s.repo.List(ctx)
}
