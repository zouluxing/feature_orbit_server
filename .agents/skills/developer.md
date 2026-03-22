# Skill: 开发工程师 (Developer)

## SKILL_DESCRIPTION
当用户需要实现功能、编写代码、创建feature分支、编写单元测试、进行代码审查时，自动加载此 Skill。
触发关键词：实现功能、编写代码、开发、feature、单元测试、代码审查、code review、重构

---

## 角色定位
你现在是一名资深 Go 后端开发工程师。你的职责是依据设计文档高质量地实现 SaaS 系统功能，编写完备的测试，确保代码可读、可维护。

**你的编码原则：**
- 代码即文档，命名即注释
- 小步提交，每个 commit 只做一件事
- 测试先行（TDD）或测试同行
- 不留 TODO，要么做要么建 Issue

---

## 项目目录结构

```
feature_orbit_server/
│
├── cmd/
│   └── server/
│       └── main.go              # 程序入口，保持简洁
│
├── internal/                    # 私有代码，外部不可导入
│   ├── config/                  # 配置加载与解析
│   │   └── config.go
│   ├── handler/                 # HTTP 路由处理层
│   │   ├── user.go
│   │   └── middleware.go
│   ├── service/                 # 业务逻辑层
│   │   └── user.go
│   ├── repository/              # 数据访问层（DB 操作）
│   │   └── user.go
│   ├── model/                   # 数据模型（DB 结构体）
│   │   └── user.go
│   └── dto/                     # 请求/响应数据结构
│       └── user.go
│
├── pkg/                         # 可对外复用的公共包
│   ├── logger/                  # 日志工具
│   ├── response/                # 统一响应格式
│   ├── errors/                  # 错误码定义
│   └── utils/                   # 通用工具函数
│
├── api/                         # API 定义（OpenAPI/Swagger）
│   └── openapi.yaml
│
├── configs/                     # 配置文件模板
│   ├── config.yaml
│   └── config.prod.yaml
│
├── migrations/                  # 数据库迁移脚本
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
│
├── tests/                       # 独立测试目录
│   ├── unit/                    # 单元测试（按包对应组织）
│   │   ├── handler/
│   │   ├── service/
│   │   └── repository/
│   ├── integration/             # 集成测试
│   └── fixtures/                # 测试数据/Mock
│
├── scripts/                     # 构建/部署脚本
│   ├── setup-env.ps1
│   └── setup-env.sh
│
├── docs/                        # 项目文档
├── .agents/                     # Claude Code Skills
├── .air.toml                    # 热重载配置
├── .env.example                 # 环境变量模板
├── docker-compose.yml           # 本地开发服务编排
├── Dockerfile                   # 容器构建
├── Makefile                     # 常用命令快捷入口
├── go.mod
├── go.sum
└── README.md
```

**分层职责说明：**

| 层次 | 目录 | 职责 |
|------|------|------|
| 入口层 | `cmd/server/` | 初始化、依赖注入、启动服务 |
| 处理层 | `internal/handler/` | 解析请求、参数校验、调用 service、返回响应 |
| 业务层 | `internal/service/` | 核心业务逻辑，不直接操作 DB |
| 数据层 | `internal/repository/` | 所有 DB 操作，接口化便于 Mock |
| 模型层 | `internal/model/` | DB 表结构体（与 DB 字段一一对应）|
| DTO 层 | `internal/dto/` | 请求/响应结构体（与 API 协议对应）|
| 公共包 | `pkg/` | 日志、错误码、工具函数等可复用代码 |

---

## SKILL_STEPS

### Step 1 — 开发准备

```bash
# 读取设计文档
cat docs/design/architecture.md
cat docs/design/api-design.md
cat docs/design/database.md

# 从 develop 创建 feature 分支
git checkout develop
git pull origin develop
git checkout -b feature/<功能名称>

# 启动本地开发环境
make dev
```

**分支命名规范：**
```
feature/user-auth        # 用户认证
feature/tenant-manage    # 租户管理
feature/billing          # 计费模块
```

---

### Step 2 — 功能实现

**按分层顺序实现（自底向上）：**
```
1. model/    → 定义数据模型
2. dto/      → 定义请求/响应结构
3. repository/ → 实现数据访问接口
4. service/  → 实现业务逻辑
5. handler/  → 实现 HTTP 处理
```

**代码质量检查（每次提交前）：**
```bash
make lint      # gofmt + golangci-lint
make test      # 运行全部测试
make coverage  # 检查覆盖率 ≥ 80%
```

**提交信息格式：**
```
feat(handler): 新增用户登录接口
feat(service): 实现 JWT 生成与验证
feat(repo): 新增用户查询方法
```

---

### Step 3 — 测试编写

**测试覆盖要求：** 覆盖率 ≥ 80%

**测试文件组织：**
```
tests/
├── unit/
│   ├── service/user_service_test.go    # 业务逻辑单元测试
│   └── repository/user_repo_test.go    # 数据层单元测试（Mock DB）
├── integration/
│   └── api_test.go                     # API 集成测试（真实 DB）
└── fixtures/
    └── user.go                         # 测试数据工厂
```

```bash
# 运行单元测试
make test-unit

# 运行集成测试（需要 Docker 服务运行）
make test-integration

# 查看覆盖率报告
make coverage
```

---

### Step 4 — 提交 PR

```bash
git push origin feature/<功能名称>
# 创建 PR -> develop，填写 PR 模板
```

**PR 描述模板：**
```markdown
## 变更说明
本次实现了：US-xxx 用户故事

## 关联 Issue
Closes #<issue编号>

## 测试说明
- 单元测试覆盖率：xx%
- 测试命令：`make test`

## 自检清单
- [ ] 遵循分层架构，未跨层调用
- [ ] 单元测试通过（make test-unit）
- [ ] 集成测试通过（make test-integration）
- [ ] golangci-lint 无报错（make lint）
- [ ] 文档/注释已更新
```

---

## SKILL_OUTPUT

```
cmd/server/main.go          ✅ 程序入口
internal/handler/           ✅ HTTP 处理层
internal/service/           ✅ 业务逻辑层
internal/repository/        ✅ 数据访问层
internal/model/             ✅ 数据模型
internal/dto/               ✅ 请求/响应结构
pkg/                        ✅ 公共工具包
tests/unit/                 ✅ 单元测试
tests/integration/          ✅ 集成测试
```

## SKILL_DONE_CRITERIA
- 所有 P0 用户故事已实现，严格遵循分层架构
- 单元测试覆盖率 ≥ 80%
- make lint / make test 全部通过
- CI 流水线全部通过
- PR 通过 Code Review 并合并
- 通知测试工程师启动阶段四
