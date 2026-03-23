.PHONY: dev build test test-unit coverage lint fmt services-up services-down db-migrate keys-gen docker-build

APP    := ums
BIN    := bin/$(APP)
CMD    := ./cmd/server
DB_URL ?= postgres://postgres:postgres@localhost:5433/ums_db?sslmode=disable

dev:          air
build:        go build -ldflags "-s -w" -o $(BIN) $(CMD)
run:          go run $(CMD)
test:         go test ./... -v -timeout 60s
test-unit:    go test ./tests/unit/... -v -timeout 30s
coverage:     go test ./... -coverprofile=coverage.out -covermode=atomic && go tool cover -html=coverage.out -o coverage.html
lint:         golangci-lint run ./...
fmt:          gofmt -w .
services-up:  docker compose up -d postgres redis && sleep 4
services-down:docker compose down
db-migrate:   migrate -path migrations -database "$(DB_URL)" up
db-rollback:  migrate -path migrations -database "$(DB_URL)" down 1
docker-build: docker build -t $(APP):latest .
keys-gen:
	@mkdir -p configs
	openssl genrsa -out configs/ums_rsa_private.pem 2048
	openssl rsa -in configs/ums_rsa_private.pem -pubout -out configs/ums_rsa_public.pem
