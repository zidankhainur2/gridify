.PHONY: help dev-db migrate-up migrate-down seed run-backend run-frontend test lint

help:
	@echo Gridify Commands:
	@echo   make dev-db        - Start PostgreSQL 17 container via Docker
	@echo   make migrate-up    - Run database migrations UP
	@echo   make migrate-down  - Rollback database migrations DOWN
	@echo   make seed          - Seed sample tournaments and organizer account
	@echo   make run-backend   - Run Go API server
	@echo   make run-frontend  - Run Next.js frontend dev server
	@echo   make test          - Run backend test suite
	@echo   make lint          - Run Go linter/vet

dev-db:
	docker compose up -d postgres

migrate-up:
	cd backend && go run ./cmd/migrate up

migrate-down:
	cd backend && go run ./cmd/migrate down

seed:
	cd backend && go run ./cmd/seed

run-backend:
	cd backend && go run ./cmd/api

run-frontend:
	cd frontend && npm run dev

test:
	cd backend && go test -v -race ./...

lint:
	cd backend && go vet ./...
