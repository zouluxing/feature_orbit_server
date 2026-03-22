# ================================================
# feature_orbit_server Makefile
# 用法: make <target>
# ================================================

.PHONY: help dev build test test-unit test-integration coverage lint clean \
        services-up services-down db-migrate db-rollback health docker-build

# 默认目标：显示帮助
help:
	@echo ""
	@echo "================================================"
	@echo "  feature_orbit_server 常用命令"
	@echo "================================================"
	@echo "  开发"
	@echo "    make dev              启动热重载开发服务器"
	@echo "    make build            编译生产二进制文件"
	@echo "    make run              直接运行（不热重载）"
	@echo ""
	@echo "  测试"
	@echo "    make test             运行全部测试"
	@echo "    make test-unit        运行单元测试"
	@echo "    make test-integration 运行集成测试"
	@echo "    make coverage         生成覆盖率报告"
	@echo ""
	@echo "  代码质量"
	@echo "    make lint             运行代码规范检查"
	@echo "    make fmt              格式化代码"
	@echo "    make vet              运行 go vet"
	@echo ""
	@echo "  数据库"
	@echo "    make db-migrate       执行数据库迁移"
	@echo "    make db-rollback      回滚最近一次迁移"
	@echo "    make db-status        查看迁移状态"
	@echo ""
	@echo "  服务"
	@echo "    make services-up      启动 PostgreSQL + Redis"
	@echo "    make services-down    停止所有服务"
	@echo "    make health           检查服务健康状态"
	@echo ""
	@echo "  Docker"
	@echo "    make docker-build     构建 Docker 镜像"
	@echo "    make docker-run       运行 Docker 容器"
	@echo "================================================"
	@echo ""

# ------------------------------------------------
# 变量定义
# ------------------------------------------------
APP_NAME    := feature_orbit_server
BIN_DIR     := bin
CMD_PATH    := ./cmd/server
GO_FILES    := $(shell find . -name '*.go' -not -path './vendor/*')
BUILD_FLAGS := -ldflags "-s -w"

# 从 .env 加载数据库 URL（Windows 兼容）
ifdef DATABASE_URL
  DB_URL := $(DATABASE_URL)
else
  DB_URL := postgres://postgres:postgres@localhost:5432/feature_orbit_server?sslmode=disable
endif

# ------------------------------------------------
# 开发
# ------------------------------------------------
dev:
	@echo "🔥 启动热重载开发服务器..."
	air

build:
	@echo "🔨 编译 $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "✅ 编译完成: $(BIN_DIR)/$(APP_NAME)"

run:
	@echo "🚀 启动服务..."
	go run $(CMD_PATH)

clean:
	@echo "🧹 清理构建产物..."
	@rm -rf $(BIN_DIR)
	@go clean

# ------------------------------------------------
# 测试
# ------------------------------------------------
test:
	@echo "🧪 运行全部测试..."
	go test ./... -v -timeout 60s

test-unit:
	@echo "🧪 运行单元测试..."
	go test ./tests/unit/... -v -timeout 30s

test-integration:
	@echo "🧪 运行集成测试..."
	go test ./tests/integration/... -v -timeout 60s

coverage:
	@echo "📊 生成覆盖率报告..."
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 报告已生成: coverage.html"

# ------------------------------------------------
# 代码质量
# ------------------------------------------------
lint:
	@echo "🔍 运行代码规范检查..."
	gofmt -l $(GO_FILES)
	go vet ./...
	golangci-lint run ./...
	@echo "✅ Lint 检查通过"

fmt:
	@echo "✨ 格式化代码..."
	gofmt -w $(GO_FILES)
	goimports -w $(GO_FILES)
	@echo "✅ 格式化完成"

vet:
	go vet ./...

# ------------------------------------------------
# 数据库迁移
# ------------------------------------------------
db-migrate:
	@echo "🗄️  执行数据库迁移..."
	migrate -path migrations -database "$(DB_URL)" up
	@echo "✅ 迁移完成"

db-rollback:
	@echo "⏪ 回滚最近一次迁移..."
	migrate -path migrations -database "$(DB_URL)" down 1

db-status:
	@echo "📋 迁移状态:"
	migrate -path migrations -database "$(DB_URL)" version

db-create:
	@read -p "迁移名称: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# ------------------------------------------------
# Docker 服务
# ------------------------------------------------
services-up:
	@echo "🐳 启动 PostgreSQL + Redis..."
	docker compose up -d postgres redis
	@echo "⏳ 等待服务就绪..."
	@sleep 3
	@echo "✅ 服务已启动"

services-down:
	@echo "🛑 停止所有服务..."
	docker compose down

services-logs:
	docker compose logs -f

health:
	@echo "💓 检查服务健康状态..."
	@docker exec feature_orbit_server_postgres pg_isready -U postgres && echo "✅ PostgreSQL 正常" || echo "❌ PostgreSQL 异常"
	@docker exec feature_orbit_server_redis redis-cli ping && echo "✅ Redis 正常" || echo "❌ Redis 异常"

# ------------------------------------------------
# Docker 镜像
# ------------------------------------------------
docker-build:
	@echo "🐳 构建 Docker 镜像..."
	docker build -t $(APP_NAME):latest .
	@echo "✅ 镜像构建完成"

docker-run:
	docker run --rm -p 8080:8080 --env-file .env $(APP_NAME):latest
