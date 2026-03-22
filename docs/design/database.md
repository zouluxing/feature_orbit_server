# UMS 数据库设计

## ER 图（Mermaid）

```mermaid
erDiagram
    users ||--o{ user_roles : has
    roles ||--o{ user_roles : assigned_to
    roles ||--o{ role_permissions : has
    permissions ||--o{ role_permissions : granted_by
    users ||--o{ oauth2_tokens : owns
    oauth2_clients ||--o{ oauth2_tokens : issues
    oauth2_clients ||--o{ oauth2_authorization_codes : generates
    users ||--o{ oauth2_clients : registers

    users {
        bigserial id PK
        uuid uuid UK
        varchar email UK
        varchar username UK
        varchar password_hash
        smallint status
        boolean mfa_enabled
        varchar mfa_secret
        int failed_login_count
        timestamptz last_login_at
        jsonb extra
        timestamptz created_at
        timestamptz deleted_at
    }

    roles {
        bigserial id PK
        varchar name UK
        text description
        boolean is_system
        timestamptz created_at
    }

    permissions {
        bigserial id PK
        varchar resource
        varchar action
        text description
    }

    oauth2_clients {
        bigserial id PK
        varchar client_id UK
        varchar client_secret_hash
        varchar name
        jsonb redirect_uris
        jsonb scopes
        jsonb grant_types
        boolean is_active
        bigint owner_user_id FK
    }
```

## 核心表结构

### users
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PK | 内部主键 |
| uuid | UUID UNIQUE | 对外暴露 ID |
| email | VARCHAR(255) UNIQUE | 登录凭证 |
| password_hash | VARCHAR(255) | bcrypt cost=12 |
| status | SMALLINT | 1=active 2=disabled 3=locked |
| mfa_enabled | BOOLEAN | TOTP 开关 |
| failed_login_count | INT | 连续失败次数 |
| deleted_at | TIMESTAMPTZ | 软删除 |

### 索引策略
| 表 | 字段 | 类型 | 目的 |
|------|------|------|------|
| users | email | UNIQUE B-tree | 登录查询 |
| users | uuid | UNIQUE B-tree | 外部 ID 查询 |
| users | status, deleted_at | 复合 | 活跃用户过滤 |
| user_roles | user_id, role_id | PRIMARY KEY | 唯一性 |
| oauth2_clients | client_id | UNIQUE | OAuth 流程 |
| oauth2_authorization_codes | code, expires_at | 复合 | 授权码验证 |
