# Food Delivery (Backend)

Backend-сервисы доставки еды на Go + PostgreSQL, разработка через Docker Compose и GitHub.

## Требования
- Go (версия берётся из `go.mod`)
- Docker + Docker Compose

## Быстрый старт (dev)

### Поднять сервисы
```bash
make up

### Логи

```bash
make logs

### Остановить и удалить контейнеры/volumes:

```bash
make down

### Применить миграции через Docker

docker compose -f deploy/docker-compose.yml run --rm migrate \
  -path=/migrations \
  -database "postgres://food:food@postgres:5432/food?sslmode=disable" \
  up


### Проверка миграции

docker compose -f deploy/docker-compose.yml run --rm migrate \
  -path=/migrations \
  -database "postgres://food:food@postgres:5432/food?sslmode=disable" \
  version

# Auth API

Базовые endpoints:
POST /v1/auth/register — регистрация (email, password, role=user|courier)
POST /v1/auth/login — логин
POST /v1/auth/refresh — обновление токенов
POST /v1/auth/logout — logout (revokes refresh tokens)
GET /v1/auth/me — данные из access token

## Примеры запросов

### Регистрация

curl -s -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user1@example.com","password":"password123","role":"user"}'

### Логин

curl -s -X POST http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user1@example.com","password":"password123"}'

### /me

curl -s http://localhost:8080/v1/auth/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"

## RBAC (Roles)

### JWT содержит role. Для защищённых роутов используется middleware:

Auth(secret) — проверяет Bearer JWT и кладёт claims в context
RequireRole("admin") — требует одну из ролей (иначе 403)
Тестовый admin-only endpoint:
GET /v1/auth/admin/ping — доступен только роли admin

## CI

GitHub Actions CI запускается на push и PR и выполняет:
gofmt check
go mod tidy check
go vet ./...
go test ./... (+ -race)
golangci-lint

## Локально

gofmt -w .
go mod tidy
go vet ./...
go test ./...

## Структура

cmd/* — точки входа сервисов
internal/auth/* — auth домен (handler/service/repo)
internal/platform/* — общие компоненты (config, db, jwt, httpmw)
deploy/ — Dockerfiles / compose
migrations/ — миграции

```bash
gofmt -w $(git ls-files '*.go')
go mod tidy
go test ./...
