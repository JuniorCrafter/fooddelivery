#!/usr/bin/env bash
set -euo pipefail

# === 0) Базовые проверки ===
command -v go >/dev/null 2>&1 || { echo "ERROR: go not found"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "ERROR: docker not found"; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "ERROR: docker compose plugin not found"; exit 1; }

# === 1) Структура директорий ===
mkdir -p cmd/api cmd/courier cmd/notifications
mkdir -p internal/platform
mkdir -p db/init
mkdir -p .github/workflows

# === 2) go.mod (без внешних зависимостей на Этапе 0) ===
if [ ! -f go.mod ]; then
  go mod init food-delivery
fi
GO_MM="$(go env GOVERSION | sed 's/^go//' | cut -d. -f1,2)"
go mod edit -go="$GO_MM"

# === 3) Платформенный слой: конфиг ===
cat > internal/platform/config.go <<'EOT'
package platform

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName     string
	HTTPPort        int
	ShutdownTimeout time.Duration

	// Stage 0 readiness is a TCP dial to PostgreSQL.
	PostgresAddr  string // host:port, e.g. postgres:5432
	CheckPostgres bool

	LogLevel string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig(defaultService string) Config {
	return Config{
		ServiceName:     getenv("SERVICE_NAME", defaultService),
		HTTPPort:        getenvInt("HTTP_PORT", 8080),
		ShutdownTimeout: getenvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		PostgresAddr:    getenv("POSTGRES_ADDR", "postgres:5432"),
		CheckPostgres:   getenvBool("CHECK_POSTGRES", true),
		LogLevel:        getenv("LOG_LEVEL", "INFO"),
	}
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getenvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getenvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
EOT

# === 4) Платформенный слой: HTTP server + health/readiness ===
cat > internal/platform/httpserver.go <<'EOT'
package platform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ReadyCheck func(ctx context.Context) error

type HealthPayload struct {
	Status  string    `json:"status"`
	Service string    `json:"service"`
	Time    time.Time `json:"time"`
}

func RunHTTP(cfg Config, register func(mux *http.ServeMux), ready ReadyCheck) error {
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	mux := http.NewServeMux()

	// Built-in endpoints for container orchestration and manual checks.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, HealthPayload{Status: "ok", Service: cfg.ServiceName, Time: time.Now().UTC()})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if ready == nil {
			writeJSON(w, http.StatusOK, HealthPayload{Status: "ready", Service: cfg.ServiceName, Time: time.Now().UTC()})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := ready(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status":  "not_ready",
				"service": cfg.ServiceName,
				"time":    time.Now().UTC(),
				"error":   err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusOK, HealthPayload{Status: "ready", Service: cfg.ServiceName, Time: time.Now().UTC()})
	})

	// Custom service routes.
	if register != nil {
		register(mux)
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           loggingMiddleware(cfg, mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("[%s] http listening on %s", cfg.ServiceName, addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Printf("[%s] shutdown signal received", cfg.ServiceName)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Printf("[%s] stopped", cfg.ServiceName)
	return nil
}

func PostgresTCPReadyCheck(cfg Config) ReadyCheck {
	if !cfg.CheckPostgres {
		return nil
	}
	return func(ctx context.Context) error {
		d := net.Dialer{Timeout: 1 * time.Second}
		conn, err := d.DialContext(ctx, "tcp", cfg.PostgresAddr)
		if err != nil {
			return fmt.Errorf("postgres not reachable (%s): %w", cfg.PostgresAddr, err)
		}
		_ = conn.Close()
		return nil
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func loggingMiddleware(cfg Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s ua=%q dur=%s", cfg.ServiceName, r.Method, r.URL.Path, r.UserAgent(), time.Since(start))
	})
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.LUTC)
}
EOT

# === 5) Сервисы (3 бинарника) ===
cat > cmd/api/main.go <<'EOT'
package main

import (
	"net/http"

	"food-delivery/internal/platform"
)

func main() {
	cfg := platform.LoadConfig("api")
	ready := platform.PostgresTCPReadyCheck(cfg)

	_ = platform.RunHTTP(cfg, func(mux *http.ServeMux) {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Food Delivery API service is running\n"))
		})
	}, ready)
}
EOT

cat > cmd/courier/main.go <<'EOT'
package main

import (
	"net/http"

	"food-delivery/internal/platform"
)

func main() {
	cfg := platform.LoadConfig("courier")
	ready := platform.PostgresTCPReadyCheck(cfg)

	_ = platform.RunHTTP(cfg, func(mux *http.ServeMux) {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Courier service is running\n"))
		})
	}, ready)
}
EOT

cat > cmd/notifications/main.go <<'EOT'
package main

import (
	"net/http"

	"food-delivery/internal/platform"
)

func main() {
	cfg := platform.LoadConfig("notifications")
	ready := platform.PostgresTCPReadyCheck(cfg)

	_ = platform.RunHTTP(cfg, func(mux *http.ServeMux) {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Notifications service is running\n"))
		})
	}, ready)
}
EOT

# === 6) БД: init SQL ===
cat > db/init/001_init.sql <<'EOT'
-- Stage 0 schema bootstrap (MVP). Later stages will add constraints and reference data.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Users: both customers and couriers. Role: user|courier|admin
CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  role text NOT NULL CHECK (role IN ('user','courier','admin')),
  created_at timestamptz NOT NULL DEFAULT now()
);

-- Courier profile/location data
CREATE TABLE IF NOT EXISTS couriers (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  is_available boolean NOT NULL DEFAULT true,
  current_lat double precision,
  current_lng double precision,
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_couriers_availability ON couriers (is_available);

-- Product catalog
CREATE TABLE IF NOT EXISTS products (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  price_cents integer NOT NULL CHECK (price_cents >= 0),
  image_url text,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_products_active ON products (is_active);

-- Orders
CREATE TABLE IF NOT EXISTS orders (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id),
  courier_id uuid REFERENCES users(id),
  status text NOT NULL CHECK (status IN ('created','accepted','on_the_way','delivered','paid')),
  total_cents integer NOT NULL CHECK (total_cents >= 0),
  delivery_address text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_courier_id_status ON orders(courier_id, status);

-- Order items
CREATE TABLE IF NOT EXISTS order_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id uuid NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id uuid NOT NULL REFERENCES products(id),
  quantity integer NOT NULL CHECK (quantity > 0),
  options jsonb NOT NULL DEFAULT '{}'::jsonb,
  price_cents integer NOT NULL CHECK (price_cents >= 0)
);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- Order status history (audit log)
CREATE TABLE IF NOT EXISTS order_status_history (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id uuid NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  status text NOT NULL,
  changed_by uuid,
  changed_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history(order_id);
EOT

# === 7) Dockerfile ===
cat > Dockerfile <<'EOT'
# Stage 0: minimal, dependency-free Go services (stdlib only)

FROM golang:1.23-alpine AS build
WORKDIR /src

COPY go.mod .
# No external modules in Stage 0, but keep the command for future stages.
RUN go mod download || true

COPY . .

ARG SERVICE=api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/app ./cmd/${SERVICE}

FROM alpine:3.20
RUN adduser -D -g '' appuser && apk add --no-cache ca-certificates tzdata
WORKDIR /
COPY --from=build /out/app /app
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app"]
EOT

# === 8) docker-compose.yml ===
cat > docker-compose.yml <<'EOT'
name: food-delivery

services:
  postgres:
    image: postgres:16-alpine
    env_file:
      - .env
    ports:
      - "${POSTGRES_PORT:-5432}:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./db/init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 5s
      timeout: 3s
      retries: 20

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports:
      - "5672:5672"
      - "15672:15672"

  api:
    build:
      context: .
      args:
        SERVICE: api
    env_file:
      - .env
    environment:
      SERVICE_NAME: api
      HTTP_PORT: "8080"
      POSTGRES_ADDR: "postgres:5432"
      CHECK_POSTGRES: "true"
      LOG_LEVEL: "INFO"
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy

  courier:
    build:
      context: .
      args:
        SERVICE: courier
    env_file:
      - .env
    environment:
      SERVICE_NAME: courier
      HTTP_PORT: "8081"
      POSTGRES_ADDR: "postgres:5432"
      CHECK_POSTGRES: "true"
      LOG_LEVEL: "INFO"
    ports:
      - "8081:8081"
    depends_on:
      postgres:
        condition: service_healthy

  notifications:
    build:
      context: .
      args:
        SERVICE: notifications
    env_file:
      - .env
    environment:
      SERVICE_NAME: notifications
      HTTP_PORT: "8082"
      POSTGRES_ADDR: "postgres:5432"
      CHECK_POSTGRES: "true"
      LOG_LEVEL: "INFO"
    ports:
      - "8082:8082"
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  pgdata:
EOT

# === 9) env + make + misc ===
cat > .env.example <<'EOT'
# PostgreSQL
POSTGRES_USER=fd
POSTGRES_PASSWORD=fdpass
POSTGRES_DB=food_delivery
POSTGRES_PORT=5432
EOT
if [ ! -f .env ]; then cp .env.example .env; fi

cat > Makefile <<'EOT'
SHELL := /usr/bin/env bash

.PHONY: help up down rebuild logs ps curl-health test

help:
	@echo "Targets:"
	@echo "  make up         - build & start all containers"
	@echo "  make down       - stop containers and remove volumes"
	@echo "  make rebuild    - rebuild images (no cache)"
	@echo "  make logs       - follow logs"
	@echo "  make ps         - show container status"
	@echo "  make curl-health - call /healthz and /readyz endpoints"
	@echo "  make test       - run unit tests (host)"

up:
	docker compose --env-file .env up -d --build

down:
	docker compose down -v

rebuild:
	docker compose --env-file .env build --no-cache

logs:
	docker compose logs -f --tail=200

ps:
	docker compose ps

curl-health:
	@echo "API /healthz" && curl -fsS http://localhost:8080/healthz && echo
	@echo "API /readyz"  && curl -fsS http://localhost:8080/readyz  && echo
	@echo "Courier /healthz" && curl -fsS http://localhost:8081/healthz && echo
	@echo "Notifications /healthz" && curl -fsS http://localhost:8082/healthz && echo

test:
	go test ./...
EOT

cat > .gitignore <<'EOT'
# Binaries
/bin/
*.exe
*.out

# Build artifacts
/dist/

# Local env
.env

# IDE
.idea/
.vscode/

# OS
.DS_Store
EOT

cat > .editorconfig <<'EOT'
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
indent_style = tab
indent_size = 4
trim_trailing_whitespace = true

[*.md]
indent_style = space
indent_size = 2
trim_trailing_whitespace = false
EOT

cat > README.md <<'EOT'
# Food Delivery Backend (Stage 0)

Stage 0 goal: bootstrapped repository, Docker Compose environment, and 3 runnable Go services.

## Quick start (Ubuntu)

1) Copy env file and start containers:

```bash
cp -n .env.example .env
make up
