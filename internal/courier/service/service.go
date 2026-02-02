package service

import (
	"context"
	"log"

	"github.com/JuniorCrafter/fooddelivery/internal/courier/repo"
)

type Service interface {
	FindWork(ctx context.Context) ([]repo.OrderInfo, error)
	// Тут тоже добавляем string в возвращаемые значения
	TakeOrder(ctx context.Context, courierID, orderID int64) (string, error)
	ChangeStatus(ctx context.Context, orderID int64, status string) error // Новое!
	GetDashboard(ctx context.Context, courierID int64) (repo.Summary, []repo.OrderInfo, error)
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

func (s *courierService) TakeOrder(ctx context.Context, courierID, orderID int64) (string, error) {
	return s.repo.AcceptOrder(ctx, courierID, orderID)
}

func (s *courierService) ChangeStatus(ctx context.Context, orderID int64, status string) error {
	// 1. Обновляем в базе (PostgreSQL), как и раньше
	err := s.repo.UpdateStatus(ctx, orderID, status)
	if err != nil {
		return err
	}

	// 2. Публикуем событие в RabbitMQ (в будущем вынесем в отдельный репозиторий)
	// Пока для теста просто логируем, но в коде здесь будет вызов Publish()
	log.Printf(" Статус заказа %d изменен на %s. Событие отправлено в очередь.", orderID, status)

	return nil
}

func (s *courierService) GetDashboard(ctx context.Context, courierID int64) (repo.Summary, []repo.OrderInfo, error) {
	summary, err := s.repo.GetCourierSummary(ctx, courierID)
	if err != nil {
		return repo.Summary{}, nil, err
	}

	history, err := s.repo.GetCourierHistory(ctx, courierID)
	return summary, history, err
}
