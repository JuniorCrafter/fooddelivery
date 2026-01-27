package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool" // Библиотека для управления пулом соединений
)

// NewPool создает "бассейн" соединений с базой.
// Это эффективнее, чем открывать новое соединение на каждый чих.
func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул: %w", err)
	}

	// Проверяем, что база реально отвечает
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("база недоступна: %w", err)
	}

	return pool, nil
}
