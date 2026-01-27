package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// User — это описание того, как пользователь выглядит в базе
type User struct {
	ID       int64
	Email    string
	Password string // Тут будет лежать уже зашифрованный пароль
}

// Repository — это список команд, которые наш "кладовщик" умеет выполнять
type Repository interface {
	CreateUser(ctx context.Context, user User) (int64, error)
	GetByEmail(ctx context.Context, email string) (User, error)
}

// pgRepo — конкретная реализация кладовщика для PostgreSQL
type pgRepo struct {
	db *pgxpool.Pool
}

// New — функция-"завод", которая создает кладовщика
func New(db *pgxpool.Pool) Repository {
	return &pgRepo{db: db}
}

func (r *pgRepo) CreateUser(ctx context.Context, u User) (int64, error) {
	var id int64
	query := "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id"

	// Выполняем SQL запрос
	err := r.db.QueryRow(ctx, query, u.Email, u.Password).Scan(&id)
	return id, err
}

func (r *pgRepo) GetByEmail(ctx context.Context, email string) (User, error) {
	var u User
	query := "SELECT id, email, password FROM users WHERE email = $1"
	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Password)
	return u, err
}
