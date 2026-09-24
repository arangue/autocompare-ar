-include .env
export

DATABASE_URL ?= postgres://user:password@localhost:5435/autocompare_db?sslmode=disable
INGEST_FILE ?= testdata/listings.json

.PHONY: help dev db-up db-down api web ingest catalog-check build fmt tidy clean

help:
	@echo "  make dev            -> start Postgres and run the API"
	@echo "  make db-up          -> start Postgres"
	@echo "  make db-down        -> stop Postgres"
	@echo "  make api            -> run API (requires DATABASE_URL)"
	@echo "  make web            -> run Next.js frontend"
	@echo "  make ingest         -> upsert listings from INGEST_FILE (default testdata/listings.json)"
	@echo "  make catalog-check  -> fail if seed trims have inverted years or empty names (needs db-up + migrations)"
	@echo "  make build          -> compile API binary to bin/server"
	@echo "  make clean          -> remove containers and volumes"

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
	cd web && PORT=3000 npm run dev

ingest:
	go run ./cmd/worker $(INGEST_FILE)

catalog-check:
	go test ./internal/infrastructure/postgres/ -run Catalog -count=1

build:
	go build -o bin/server ./cmd/server

fmt:
	go fmt ./...

tidy:
	go mod tidy

clean:
	docker compose down -v
