package handler

import (
	"encoding/json"
	"net/http"

	"github.com/JuniorCrafter/fooddelivery/internal/auth/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	authService service.Service
}

func New(s service.Service) *Handler {
	return &Handler{authService: s}
}

// RegisterRoutes "рисует" карту наших ручек (маршруты)
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login) // Новая ручка!
}

// Структура для чтения данных из JSON запроса
type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	// Читаем JSON из запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат данных", http.StatusBadRequest)
		return
	}

	// Вызываем наш сервис
	id, err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем ответ, что всё ок
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"user_id": id})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req registerRequest // используем ту же структуру для email и password

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неверный формат", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Отправляем токен пользователю
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
