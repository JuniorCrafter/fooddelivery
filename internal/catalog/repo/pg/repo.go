package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Product struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
}

type Repository interface {
	Create(ctx context.Context, p Product) (int64, error)
	List(ctx context.Context) ([]Product, error)
}

type pgRepo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Repository {
	return &pgRepo{db: db}
}

func (r *pgRepo) Create(ctx context.Context, p Product) (int64, error) {
	var id int64
	query := "INSERT INTO products (name, description, price, image_url) VALUES ($1, $2, $3, $4) RETURNING id"
	err := r.db.QueryRow(ctx, query, p.Name, p.Description, p.Price, p.ImageURL).Scan(&id)
	return id, err
}

func (r *pgRepo) List(ctx context.Context) ([]Product, error) {
	query := "SELECT id, name, description, price, image_url FROM products"
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}
