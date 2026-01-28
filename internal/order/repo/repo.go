package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderItem struct {
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Order struct {
	ID         int64       `json:"id"`
	UserID     int64       `json:"user_id"`
	TotalPrice float64     `json:"total_price"`
	Items      []OrderItem `json:"items"`
}

type Repository interface {
	CreateOrder(ctx context.Context, o Order) (int64, error)
}

type pgRepo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Repository {
	return &pgRepo{db: db}
}

func (r *pgRepo) CreateOrder(ctx context.Context, o Order) (int64, error) {
	// Начинаем транзакцию
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) // Если что-то пойдет не так, изменения откатятся

	// 1. Создаем сам заказ
	var orderID int64
	err = tx.QueryRow(ctx, "INSERT INTO orders (user_id, total_price) VALUES ($1, $2) RETURNING id", o.UserID, o.TotalPrice).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	// 2. Добавляем каждый товар из заказа
	for _, item := range o.Items {
		_, err = tx.Exec(ctx, "INSERT INTO order_items (order_id, product_id, quantity, price_at_purchase) VALUES ($1, $2, $3, $4)",
			orderID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return 0, err
		}
	}

	// Подтверждаем транзакцию
	return orderID, tx.Commit(ctx)
}
