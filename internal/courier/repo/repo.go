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

type Summary struct {
	TotalOrders   int     `json:"total_orders"`
	TotalEarnings float64 `json:"total_earnings"`
}

type Repository interface {
	GetNewOrders(ctx context.Context) ([]OrderInfo, error)
	// Важно: здесь (string, error)
	AcceptOrder(ctx context.Context, courierID, orderID int64) (string, error)
	UpdateStatus(ctx context.Context, orderID int64, status string) error
	GetCourierHistory(ctx context.Context, courierID int64) ([]OrderInfo, error)
	GetCourierSummary(ctx context.Context, courierID int64) (Summary, error)
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

func (r *pgRepo) GetCourierHistory(ctx context.Context, courierID int64) ([]OrderInfo, error) {
	query := "SELECT id, total_price FROM orders WHERE courier_id = $1 AND status = 'completed' ORDER BY created_at DESC"
	rows, err := r.db.Query(ctx, query, courierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []OrderInfo
	for rows.Next() {
		var o OrderInfo
		rows.Scan(&o.ID, &o.TotalPrice)
		history = append(history, o)
	}
	return history, nil
}

func (r *pgRepo) GetCourierSummary(ctx context.Context, courierID int64) (Summary, error) {
	var s Summary
	// Считаем: (количество заказов * 100р) + (10% от суммы всех заказов)
	query := `
		SELECT
			COUNT(id),
			COALESCE(SUM(total_price * 0.1 + 100), 0)
		FROM orders
		WHERE courier_id = $1 AND status = 'completed'`

	err := r.db.QueryRow(ctx, query, courierID).Scan(&s.TotalOrders, &s.TotalEarnings)
	return s, err
}
