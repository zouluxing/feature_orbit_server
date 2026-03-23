# 部署运维手册

## 服务拓扑

```
[Nginx]
  ├──► feature_orbit_server:8080 ─► PostgreSQL:5432
  └──► UMS:8081 ─► PostgreSQL:5433 + Redis:6380
```

## 启动顺序

1. PostgreSQL (ums_db + feature_orbit_db)
2. Redis
3. UMS
4. feature_orbit_server

## 关键环境变量

### UMS
| 变量 | 默认值 |
|------|--------|
| `UMS_DB_URL` | — |
| `UMS_REDIS_ADDR` | localhost:6380 |
| `UMS_JWT_PRIVATE_KEY_PATH` | configs/ums_rsa_private.pem |
| `UMS_MFA_ENCRYPTION_KEY` | 生产必填 (openssl rand -hex 32) |
| `UMS_APP_PORT` | 8081 |

### feature_orbit_server
| 变量 | 默认值 |
|------|--------|
| `APP_PORT` | 8080 |
| `APP_UMS_BASE_URL` | http://localhost:8081 |
| `APP_UMS_CACHE_TTL` | 1h |

## 密钥管理

```bash
# 生成 RSA 密钥对
openssl genrsa -out ums_rsa_private.pem 2048
openssl rsa -in ums_rsa_private.pem -pubout -out ums_rsa_public.pem

# 生成 MFA 加密密钥
openssl rand -hex 32
```

## 健康检查

```bash
curl http://localhost:8081/health
curl http://localhost:8081/.well-known/jwks.json
curl http://localhost:8080/health
```

## 故障排查

- UMS 启动失败: 检查 PostgreSQL/Redis 连通性、RSA 密钥文件权限
- JWT 验证失败: 检查 JWKS 缓存 TTL（默认 1h）、密钥轮换后刷新
- MFA 验证失败: 检查 UMS_MFA_ENCRYPTION_KEY 是否一致
