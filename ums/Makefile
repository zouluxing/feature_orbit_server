# ================================================
# UMS Makefile
# ================================================

.PHONY: help dev build run test test-unit coverage lint fmt \
        services-up services-down health db-migrate db-rollback \
        docker-build keys-gen

APP      := ums
BIN      := bin/$(APP)
CMD      := ./cmd/server
DB_URL   ?= postgres://postgres:postgres@localhost:5433/ums_db?sslmode=disable

help:
	@echo ""
	@echo "  UMS — User Management Service"
	@echo "  make dev           hot-reload server (air)"
	@echo "  make build         compile binary"
	@echo "  make test          run all tests"
	@echo "  make test-unit     run unit tests only"
	@echo "  make coverage      generate HTML coverage report"
	@echo "  make lint          golangci-lint"
	@echo "  make services-up   start postgres + redis (Docker)"
	@echo "  make db-migrate    run migrations up"
	@echo "  make docker-build  build Docker image"
	@echo ""

dev:
	air

build:
	@mkdir -p bin
	go build -ldflags "-s -w" -o $(BIN) $(CMD)

run:
	go run $(CMD)

test:
	go test ./... -v -timeout 60s

test-unit:
	go test ./tests/unit/... -v -timeout 30s

coverage:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out | grep total
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

services-up:
	docker compose up -d postgres redis
	@sleep 4

services-down:
	docker compose down

health:
	@docker exec ums_postgres pg_isready -U postgres
	@docker exec ums_redis redis-cli -a redis_password ping

db-migrate:
	migrate -path migrations -database "$(DB_URL)" up

db-rollback:
	migrate -path migrations -database "$(DB_URL)" down 1

db-status:
	migrate -path migrations -database "$(DB_URL)" version

docker-build:
	docker build -t $(APP):latest .

keys-gen:
	@mkdir -p configs
	openssl genrsa -out configs/ums_rsa_private.pem 2048
	openssl rsa -in configs/ums_rsa_private.pem -pubout -out configs/ums_rsa_public.pem
