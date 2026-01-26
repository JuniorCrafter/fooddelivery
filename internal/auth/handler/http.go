package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/JuniorCrafter/fooddelivery/internal/auth/repo"
	"github.com/JuniorCrafter/fooddelivery/internal/auth/service"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/httpmw"
)

type Handler struct {
	svc       *service.Service
	jwtSecret []byte
}

func New(svc *service.Service, jwtSecret []byte) *Handler {
	return &Handler{svc: svc, jwtSecret: jwtSecret}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/v1/auth/register", h.register)
	r.Post("/v1/auth/login", h.login)
	r.Post("/v1/auth/refresh", h.refresh)

	r.Group(func(pr chi.Router) {
		pr.Use(httpmw.Auth(h.jwtSecret))

		pr.Get("/v1/auth/me", h.me)
		pr.Post("/v1/auth/logout", h.logout)

		// Пример RBAC-эндпоинта (только admin)
		pr.With(httpmw.RequireRole("admin")).Get("/v1/auth/admin/ping", h.adminPing)
	})

	return r
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"` // user | courier
}

type tokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	access, refresh, err := h.svc.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, repo.ErrEmailTaken) {
			http.Error(w, "email already registered", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, tokenResp{AccessToken: access, RefreshToken: refresh})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	access, refresh, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, tokenResp{AccessToken: access, RefreshToken: refresh})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	access, refresh, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, "invalid refresh", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, tokenResp{AccessToken: access, RefreshToken: refresh})
}

type meResp struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpmw.Claims(r.Context())
	if !ok {
		http.Error(w, "missing claims", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, meResp{UserID: claims.UserID, Role: claims.Role})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpmw.Claims(r.Context())
	if !ok {
		http.Error(w, "missing claims", http.StatusUnauthorized)
		return
	}
	if err := h.svc.Logout(r.Context(), claims.UserID); err != nil {
		http.Error(w, "logout failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "failed to encode json", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if _, err := w.Write(append(data, '\n')); err != nil {
		log.Printf("write response: %v", err)
	}
}

func (h *Handler) adminPing(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
