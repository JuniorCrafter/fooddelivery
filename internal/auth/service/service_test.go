package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JuniorCrafter/fooddelivery/internal/auth/repo"
)

type fakeRepo struct {
	usersByEmail map[string]repo.User
	usersByID    map[int64]repo.User

	nextUserID int64

	// refresh tokens: rawToken -> userID
	refresh map[string]int64

	// knobs
	createUserErr error
	getByEmailErr error
	getByIDErr    error
	useRefreshErr error
	revokeErr     error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		usersByEmail: make(map[string]repo.User),
		usersByID:    make(map[int64]repo.User),
		nextUserID:   1,
		refresh:      make(map[string]int64),
	}
}

func (f *fakeRepo) CreateUser(ctx context.Context, email, passHash, role string) (int64, error) {
	if f.createUserErr != nil {
		return 0, f.createUserErr
	}
	if _, ok := f.usersByEmail[email]; ok {
		return 0, repo.ErrEmailTaken
	}
	id := f.nextUserID
	f.nextUserID++

	u := repo.User{ID: id, Email: email, PasswordHash: passHash, Role: role}
	f.usersByEmail[email] = u
	f.usersByID[id] = u
	return id, nil
}

func (f *fakeRepo) GetUserByEmail(ctx context.Context, email string) (repo.User, error) {
	if f.getByEmailErr != nil {
		return repo.User{}, f.getByEmailErr
	}
	u, ok := f.usersByEmail[email]
	if !ok {
		return repo.User{}, errors.New("not found")
	}
	return u, nil
}

func (f *fakeRepo) GetUserByID(ctx context.Context, id int64) (repo.User, error) {
	if f.getByIDErr != nil {
		return repo.User{}, f.getByIDErr
	}
	u, ok := f.usersByID[id]
	if !ok {
		return repo.User{}, errors.New("not found")
	}
	return u, nil
}

func (f *fakeRepo) SaveRefreshToken(ctx context.Context, userID int64, rawToken string, expiresAt time.Time) error {
	// для unit-тестов достаточно, что токен сохраняется
	f.refresh[rawToken] = userID
	return nil
}

func (f *fakeRepo) UseRefreshToken(ctx context.Context, rawToken string) (int64, error) {
	if f.useRefreshErr != nil {
		return 0, f.useRefreshErr
	}
	uid, ok := f.refresh[rawToken]
	if !ok {
		return 0, errors.New("invalid refresh")
	}
	// одноразовость: удаляем после использования
	delete(f.refresh, rawToken)
	return uid, nil
}

func (f *fakeRepo) RevokeAllRefreshTokens(ctx context.Context, userID int64) error {
	if f.revokeErr != nil {
		return f.revokeErr
	}
	// простая реализация: вычищаем все токены данного пользователя
	for tok, uid := range f.refresh {
		if uid == userID {
			delete(f.refresh, tok)
		}
	}
	return nil
}

func TestRegister_Success(t *testing.T) {
	r := newFakeRepo()
	s := New(r, []byte("secret"), 15*time.Minute, 30*24*time.Hour, 4)

	access, refresh, err := s.Register(context.Background(), "user1@example.com", "password123", "user")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("expected non-empty tokens")
	}
	u, err := r.GetUserByEmail(context.Background(), "user1@example.com")
	if err != nil {
		t.Fatalf("expected user created, got %v", err)
	}
	if u.Role != "user" {
		t.Fatalf("expected role user, got %s", u.Role)
	}
}

func TestRegister_EmailTaken(t *testing.T) {
	r := newFakeRepo()
	s := New(r, []byte("secret"), 15*time.Minute, 30*24*time.Hour, 4)

	_, _, err := s.Register(context.Background(), "user1@example.com", "password123", "user")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	_, _, err = s.Register(context.Background(), "user1@example.com", "password123", "user")
	if !errors.Is(err, repo.ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegister_InvalidInput(t *testing.T) {
	r := newFakeRepo()
	s := New(r, []byte("secret"), 15*time.Minute, 30*24*time.Hour, 4)

	_, _, err := s.Register(context.Background(), "", "password123", "user")
	if err == nil {
		t.Fatalf("expected error for empty email")
	}

	_, _, err = s.Register(context.Background(), "a@b.com", "short", "user")
	if err == nil {
		t.Fatalf("expected error for short password")
	}
}

func TestLogin_Success(t *testing.T) {
	r := newFakeRepo()
	s := New(r, []byte("secret"), 15*time.Minute, 30*24*time.Hour, 4)

	_, _, err := s.Register(context.Background(), "user1@example.com", "password123", "user")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	access, refresh, err := s.Login(context.Background(), "user1@example.com", "password123")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatalf("expected non-empty tokens")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	r := newFakeRepo()
	s := New(r, []byte("secret"), 15*time.Minute, 30*24*time.Hour, 4)

	_, _, err := s.Register(context.Background(), "user1@example.com", "password123", "user")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, _, err = s.Login(context.Background(), "user1@example.com", "wrongpass")
	if err == nil {
		t.Fatalf("expected error for wrong password")
	}
}

func TestRefresh_Success(t *testing.T) {
	r := newFakeRepo()
	s := New(r, []byte("secret"), 15*time.Minute, 30*24*time.Hour, 4)

	_, refresh, err := s.Register(context.Background(), "user1@example.com", "password123", "user")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	newAccess, newRefresh, err := s.Refresh(context.Background(), refresh)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if newAccess == "" || newRefresh == "" {
		t.Fatalf("expected non-empty tokens")
	}
}

func TestLogout_RevokesTokens(t *testing.T) {
	r := newFakeRepo()
	s := New(r, []byte("secret"), 15*time.Minute, 30*24*time.Hour, 4)

	_, refresh, err := s.Register(context.Background(), "user1@example.com", "password123", "user")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	u, _ := r.GetUserByEmail(context.Background(), "user1@example.com")
	if err := s.Logout(context.Background(), u.ID); err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	// токен должен быть отозван
	_, _, err = s.Refresh(context.Background(), refresh)
	if err == nil {
		t.Fatalf("expected error after logout")
	}
}
