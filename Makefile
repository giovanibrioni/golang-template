.DEFAULT_GOAL := help

.PHONY: help install up down dev migrate test-unit test lint

help: ## Show available targets
	@echo "Targets:"
	@echo "  install     Download Go modules"
	@echo "  up          Start Postgres and ensure app_test exists"
	@echo "  down        Stop Postgres"
	@echo "  dev         Run the API on :8081"
	@echo "  migrate     Apply migrations and exit"
	@echo "  test-unit   Run tests that do not need Postgres"
	@echo "  test        Run all tests against app_test"
	@echo "  lint        Check formatting and go vet"

install:
	go mod download

up:
	docker compose up -d --wait
	docker compose exec -T db psql -U postgres -tc \
	  "SELECT 1 FROM pg_database WHERE datname = 'app_test'" \
	  | grep -q 1 || docker compose exec -T db psql -U postgres -c "CREATE DATABASE app_test"

down:
	docker compose down

dev:
	DATABASE_URL=postgres://postgres:postgres@localhost:5432/app?sslmode=disable \
	  go run ./cmd/api

migrate:
	DATABASE_URL=postgres://postgres:postgres@localhost:5432/app?sslmode=disable \
	  go run ./cmd/api -migrate-only

test-unit:
	go test -short ./...

test:
	DATABASE_URL=postgres://postgres:postgres@localhost:5432/app_test?sslmode=disable \
	  go test ./...

lint:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
	  printf '%s\n' "$$files"; \
	  exit 1; \
	fi
	go vet ./...
