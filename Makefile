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
	@echo "API /healthz" && curl -fsS http://localhost:18080/healthz && echo
	@echo "API /readyz"  && curl -fsS http://localhost:18080/readyz  && echo
	@echo "Courier /healthz" && curl -fsS http://localhost:18081/healthz && echo
	@echo "Notifications /healthz" && curl -fsS http://localhost:18082/healthz && echo

test:
	go test ./...

COMPOSE = docker compose -f deploy/docker-compose.yml

.PHONY: up down logs ps
up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f --tail=200

ps:
	$(COMPOSE) ps

COMPOSE = docker compose -f deploy/docker-compose.yml

.PHONY: up down logs ps migrate-up migrate-down migrate-version

migrate-up:
	$(COMPOSE) run --rm migrate -path=/migrations -database "$$DATABASE_URL" up

migrate-down:
	$(COMPOSE) run --rm migrate -path=/migrations -database "$$DATABASE_URL" down 1

migrate-version:
	$(COMPOSE) run --rm migrate -path=/migrations -database "$$DATABASE_URL" version

