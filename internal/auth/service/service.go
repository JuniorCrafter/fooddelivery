package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	jwtutil "food-delivery/internal/platform/jwt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/JuniorCrafter/fooddelivery/internal/auth/repo"
)

type Service struct {
	repo       *repo.Repo
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	bcryptCost int
}

func New(r *repo.Repo, jwtSecret []byte, accessTTL, refreshTTL time.Duration, bcryptCost int) *Service {
	return &Service{
		repo: r, jwtSecret: jwtSecret,
		accessTTL: accessTTL, refreshTTL: refreshTTL,
		bcryptCost: bcryptCost,
	}
}

func normalizeEmail(s string) string { return strings.TrimSpace(strings.ToLower(s)) }

func (s *Service) Register(ctx context.Context, email, password, role string) (access, refresh string, err error) {
	email = normalizeEmail(email)
	if email == "" || len(password) < 8 {
		return "", "", errors.New("invalid email or password")
	}
	if role == "" {
		role = "user"
	}
	if role != "user" && role != "courier" { // admin — только ручное создание/сидирование
		return "", "", errors.New("invalid role")
	}

	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return "", "", err
	}

	uid, err := s.repo.CreateUser(ctx, email, string(passHashBytes), role)
	if err != nil {
		return "", "", err
	}

	access, err = jwtutil.MintHS256(s.jwtSecret, uid, role, s.accessTTL)
	if err != nil {
		return "", "", err
	}
	refresh, err = s.mintRefresh(ctx, uid)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (access, refresh string, err error) {
	email = normalizeEmail(email)
	u, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}
	access, err = jwtutil.MintHS256(s.jwtSecret, u.ID, u.Role, s.accessTTL)
	if err != nil {
		return "", "", err
	}
	refresh, err = s.mintRefresh(ctx, u.ID)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (newAccess, newRefresh string, err error) {
	uid, err := s.repo.UseRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}
	// Роль нужна для access token → вытаскиваем email/role проще отдельным запросом по uid на следующем шаге.
	// Для этапа 1 добавим маленький запрос:
	u, err := s.repoGetUserByID(ctx, uid)
	if err != nil {
		return "", "", err
	}

	newAccess, err = jwtutil.MintHS256(s.jwtSecret, uid, u.Role, s.accessTTL)
	if err != nil {
		return "", "", err
	}
	newRefresh, err = s.mintRefresh(ctx, uid)
	if err != nil {
		return "", "", err
	}
	return newAccess, newRefresh, nil
}

func (s *Service) Logout(ctx context.Context, userID int64) error {
	return s.repo.RevokeAllRefreshTokens(ctx, userID)
}

func (s *Service) mintRefresh(ctx context.Context, userID int64) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(b)
	exp := time.Now().Add(s.refreshTTL)
	if err := s.repo.SaveRefreshToken(ctx, userID, raw, exp); err != nil {
		return "", err
	}
	return raw, nil
}
