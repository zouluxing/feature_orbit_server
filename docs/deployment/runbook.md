# 部署运维手册

## 服务拓扑

```
互联网
  │
  ▼
[Nginx / API Gateway]
  │
  ├──► feature_orbit_server:8080 ─────► PostgreSQL:5432 (feature_orbit_db)
  │                                   └───► [UMS 健康检查仅启动时]
  │
  └──► UMS:8081 ────────────────► PostgreSQL:5433 (ums_db)
                                    └───► Redis:6380
```

## 服务启动顺序

```
1. PostgreSQL (ums_db + feature_orbit_db)
2. Redis
3. UMS
4. feature_orbit_server
```

## 关键环境变量

### UMS

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `UMS_DB_URL` | PostgreSQL 连接串 | — |
| `UMS_REDIS_ADDR` | Redis 地址 | localhost:6380 |
| `UMS_REDIS_PASSWORD` | Redis 密码 | — |
| `UMS_JWT_PRIVATE_KEY_PATH` | RSA 私钥路径 | configs/ums_rsa_private.pem |
| `UMS_JWT_ACCESS_TTL` | access_token 有效期 | 15m |
| `UMS_JWT_REFRESH_TTL` | refresh_token 有效期 | 168h |
| `UMS_APP_PORT` | 监听端口 | 8081 |
| `UMS_APP_ENV` | 环境（development/production）| development |

### feature_orbit_server

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `APP_PORT` | 监听端口 | 8080 |
| `APP_DATABASE_URL` | PostgreSQL 连接串 | — |
| `APP_UMS_BASE_URL` | UMS 服务地址 | http://localhost:8081 |
| `APP_UMS_CACHE_TTL` | JWKS 缓存 TTL | 1h |

## 健康检查

```bash
# UMS
curl http://localhost:8081/health
# {"status":"ok","service":"ums"}

# JWKS
curl http://localhost:8081/.well-known/jwks.json
# {"keys":[{"kty":"RSA",...}]}

# feature_orbit_server
curl http://localhost:8080/health
# {"status":"ok","service":"feature_orbit_server"}
```

## RSA 密钥管理

```bash
# 生成生产密钥对（仅执行一次，安全存储）
openssl genrsa -out ums_rsa_private.pem 2048
openssl rsa -in ums_rsa_private.pem -pubout -out ums_rsa_public.pem

# 生产环境建议使用 Docker Secret 或 Vault 注入
# 不得将私钥提交到 Git 仓库
```

## 数据库迁移

```bash
# 安装 golang-migrate
brew install golang-migrate  # macOS

# 执行迁移
migrate -path ums/migrations \
  -database "postgres://postgres:postgres@localhost:5433/ums_db?sslmode=disable" up

# 回滚
migrate -path ums/migrations \
  -database "postgres://..." down 1
```

## 常见故障排查

### UMS 启动失败
1. 检查 PostgreSQL 是否就绪：`pg_isready -h localhost -p 5433`
2. 检查 Redis 连通性：`redis-cli -h localhost -p 6380 ping`
3. 检查 RSA 密钥文件是否存在且权限为 `0600`

### JWT 验证失败（主服务 401）
1. 检查 feature_orbit_server 能否访问 UMS JWKS 端点
2. 检查 JWKS 缓存是否已过期（默认 1h），重启服务强制刷新
3. 检查 UMS 是否使用了与主服务缓存不同的密钥（密钥轮换场景）

### Token 已失效但仍可用
1. 检查 Redis 黑名单是否正常写入
2. 检查主服务是否有独立的黑名单检查逻辑（当前版本不做，依赖 access_token TTL）

## 密钥轮换流程

1. 生成新密钥对
2. 更新 UMS 配置（新 `key_id`）并重启 UMS
3. UMS 新发 token 使用新密钥，旧 token 在 TTL 内自然过期
4. 主服务 JWKS 缓存 1h 内自动获取新公钥
5. 总切换时间 ≤ access_token TTL (15min) + JWKS TTL (1h)
