include .env
export

.PHONY: help dev db-up db-down api web fmt tidy clean

help:
	@echo "  make dev     -> start Postgres and run the API"
	@echo "  make db-up   -> start Postgres"
	@echo "  make db-down -> stop Postgres"
	@echo "  make api     -> run API (requires DATABASE_URL)"
	@echo "  make web     -> run Next.js frontend"
	@echo "  make clean   -> remove containers and volumes"

dev: db-up
	@sleep 2
	go run ./cmd/server

db-up:
	docker compose up -d autocompare-db

db-down:
	docker compose down

api:
	go run ./cmd/server

web:
	cd web && npm run dev

fmt:
	go fmt ./...

tidy:
	go mod tidy

clean:
	docker compose down -v
