include .env
export

MIGRATIONS_DIR := db/migrations
GOOSE := go tool goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)"

.PHONY: run dev build test lint fmt tidy \
	migrate-up migrate-down migrate-status migrate-reset migrate-create sqlc

run:
	go run ./cmd/api

dev: ## Hot reload (requires: go install github.com/air-verse/air@latest)
	air

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -race -cover

lint: ## Requires golangci-lint
	golangci-lint run ./...

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

# ---------- Database ----------

migrate-up:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-status:
	$(GOOSE) status

migrate-reset:
	$(GOOSE) reset

migrate-create: ## Usage: make migrate-create name=create_users
	@test -n "$(name)" || (echo "usage: make migrate-create name=<name>" && exit 1)
	go tool goose -dir $(MIGRATIONS_DIR) -s create $(name) sql

sqlc: ## Generate Go code from db/queries
	@if ls db/queries/*.sql >/dev/null 2>&1; then go tool sqlc generate; \
	else echo "sqlc: no queries in db/queries yet, skipping"; fi
