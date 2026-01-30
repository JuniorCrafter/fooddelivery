package repo

import (
	"context"
	"fmt" // Нужен для создания ошибок через fmt.Errorf

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderInfo struct {
	ID         int64   `json:"id"`
	TotalPrice float64 `json:"total_price"`
}

type Repository interface {
	GetNewOrders(ctx context.Context) ([]OrderInfo, error)
	// Важно: здесь (string, error)
	AcceptOrder(ctx context.Context, courierID, orderID int64) (string, error)
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
		if err := rows.Scan(&o.ID, &o.TotalPrice); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// Проверь, чтобы заголовок этой функции в точности совпадал с интерфейсом выше!
func (r *pgRepo) AcceptOrder(ctx context.Context, courierID, orderID int64) (string, error) {
	// 1. Пытаемся закрепить курьера за заказом
	queryUpdate := "UPDATE orders SET courier_id = $1, status = 'accepted' WHERE id = $2 AND courier_id IS NULL"
	result, err := r.db.Exec(ctx, queryUpdate, courierID, orderID)
	if err != nil {
		return "", err
	}

	// Если никто не обновился, значит заказ уже кто-то перехватил
	if result.RowsAffected() == 0 {
		return "", fmt.Errorf("заказ уже занят другим курьером")
	}

	// 2. Достаем имя курьера, чтобы вернуть его в ответе
	var courierName string
	queryName := "SELECT name FROM couriers WHERE id = $1"
	err = r.db.QueryRow(ctx, queryName, courierID).Scan(&courierName)
	if err != nil {
		return "", err
	}

	return courierName, nil
}

func (r *pgRepo) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	query := "UPDATE orders SET status = $1 WHERE id = $2"
	_, err := r.db.Exec(ctx, query, status, orderID)
	return err
}
