package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderInfo struct {
	ID         int64   `json:"id"`
	TotalPrice float64 `json:"total_price"`
}

type Repository interface {
	GetNewOrders(ctx context.Context) ([]OrderInfo, error)
	AcceptOrder(ctx context.Context, courierID, orderID int64) error
	UpdateStatus(ctx context.Context, orderID int64, status string) error
}

type pgRepo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Repository {
	return &pgRepo{db: db}
}

func (r *pgRepo) GetNewOrders(ctx context.Context) ([]OrderInfo, error) {
	query := "SELECT id, total_price FROM orders WHERE courier_id IS NULL AND status = 'new'"
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []OrderInfo
	for rows.Next() {
		var o OrderInfo
		rows.Scan(&o.ID, &o.TotalPrice)
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *pgRepo) AcceptOrder(ctx context.Context, courierID, orderID int64) error {
	// Привязываем курьера к заказу и меняем статус
	query := "UPDATE orders SET courier_id = $1, status = 'accepted' WHERE id = $2 AND courier_id IS NULL"
	_, err := r.db.Exec(ctx, query, courierID, orderID)
	return err
}

func (r *pgRepo) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	query := "UPDATE orders SET status = $1 WHERE id = $2"
	_, err := r.db.Exec(ctx, query, status, orderID)
	return err
}
