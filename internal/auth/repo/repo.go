package repo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
}

func (r *Repo) CreateUser(ctx context.Context, email, passHash, role string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users(email, password_hash, role) VALUES ($1,$2,$3) RETURNING id`,
		email, passHash, role,
	).Scan(&id)
	return id, err
}

func (r *Repo) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, role FROM users WHERE email=$1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)
	return u, err
}

func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (r *Repo) SaveRefreshToken(ctx context.Context, userID int64, rawToken string, expiresAt time.Time) error {
	h := hashRefreshToken(rawToken)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens(user_id, token_hash, expires_at) VALUES ($1,$2,$3)`,
		userID, h, expiresAt,
	)
	return err
}

func (r *Repo) UseRefreshToken(ctx context.Context, rawToken string) (int64, error) {
	h := hashRefreshToken(rawToken)

	// “одноразовость”: помечаем как revoked при использовании
	var userID int64
	var revokedAt *time.Time
	var expiresAt time.Time

	err := r.pool.QueryRow(ctx,
		`SELECT user_id, revoked_at, expires_at
		 FROM refresh_tokens
		 WHERE token_hash=$1`,
		h,
	).Scan(&userID, &revokedAt, &expiresAt)
	if err != nil {
		return 0, err
	}
	if revokedAt != nil {
		return 0, errors.New("refresh token revoked")
	}
	if time.Now().After(expiresAt) {
		return 0, errors.New("refresh token expired")
	}

	_, err = r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at=now() WHERE token_hash=$1`,
		h,
	)
	return userID, err
}

func (r *Repo) RevokeAllRefreshTokens(ctx context.Context, userID int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`,
		userID,
	)
	return err
}

func (r *Repo) GetUserByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, role FROM users WHERE id=$1`,
		id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)
	return u, err
}

var ErrNotFound = pgx.ErrNoRows
