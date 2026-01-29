package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/JuniorCrafter/fooddelivery/internal/auth/repo"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/jwt"
	"golang.org/x/crypto/bcrypt" // Библиотека для шифрования
)

// Service описывает, что наш сервис умеет делать
type Service interface {
	Register(ctx context.Context, email, password, role string) (int64, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type authService struct {
	repo repo.Repository
}

// New создает новый сервис, которому для работы нужен "кладовщик" (repo)
func New(r repo.Repository) Service {
	return &authService{repo: r}
}

func (s *authService) Register(ctx context.Context, email, password, role string) (int64, error) {
	// 1. Шифруем пароль. Cost 10 — это оптимальная сложность шифрования
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return 0, fmt.Errorf("ошибка шифрования: %w", err)
	}

	// 2. Создаем объект пользователя для сохранения
	u := repo.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	// 3. Просим репозиторий сохранить его в базу
	id, err := s.repo.CreateUser(ctx, u)
	if err != nil {
		return 0, fmt.Errorf("не удалось создать пользователя: %w", err)
	}

	return id, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	// 1. Ищем пользователя в базе по email
	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("пользователь не найден")
	}

	// 2. Сравниваем введенный пароль с хешем из базы
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return "", errors.New("неверный пароль")
	}

	// 3. Если всё ок — создаем токен. Пока по умолчанию роль "client"
	token, err := jwt.GenerateToken(u.ID, "client")
	return token, err
}
