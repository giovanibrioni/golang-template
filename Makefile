.DEFAULT_GOAL := help

.PHONY: help install up down dev migrate swagger test-unit test lint review

help: ## Show available targets
	@echo "Targets:"
	@echo "  install     Download Go modules"
	@echo "  up          Start Postgres and ensure app_test exists"
	@echo "  down        Stop Postgres"
	@echo "  dev         Run the API on :8081"
	@echo "  migrate     Apply migrations and exit"
	@echo "  swagger     Generate OpenAPI docs (swagger.json/yaml)"
	@echo "  test-unit   Run tests that do not need Postgres"
	@echo "  test        Run all tests against app_test"
	@echo "  lint        Check formatting and go vet"
	@echo "  review      Architecture, orphan, env, and merge-risk checks"

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
	OTEL_EXPORTER_OTLP_ENDPOINT= \
	DATABASE_URL=postgres://postgres:postgres@localhost:5432/app?sslmode=disable \
	  go run ./cmd/api

migrate:
	DATABASE_URL=postgres://postgres:postgres@localhost:5432/app?sslmode=disable \
	  go run ./cmd/api -migrate-only

swagger:
	go install github.com/swaggo/swag/cmd/swag@latest
	~/go/bin/swag init -g cmd/api/main.go -o docs

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

review:
	@status=0; \
	go run .cursor/skills/golang-template-review/scripts/architecture_check.go || status=1; \
	go run .cursor/skills/golang-template-review/scripts/orphan_check.go || status=1; \
	go run .cursor/skills/golang-template-review/scripts/env_example_check.go || status=1; \
	go run .cursor/skills/golang-template-review/scripts/merge_risk_check.go || status=1; \
	exit $$status
