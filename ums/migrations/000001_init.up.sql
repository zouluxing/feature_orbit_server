-- UMS initial schema
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id                  BIGSERIAL PRIMARY KEY,
    uuid                UUID        NOT NULL DEFAULT gen_random_uuid(),
    email               VARCHAR(255) NOT NULL,
    username            VARCHAR(64)  NOT NULL,
    password_hash       VARCHAR(255) NOT NULL,
    phone               VARCHAR(32),
    avatar_url          TEXT,
    status              SMALLINT    NOT NULL DEFAULT 1,
    mfa_enabled         BOOLEAN     NOT NULL DEFAULT FALSE,
    mfa_secret          VARCHAR(64),
    failed_login_count  INT         NOT NULL DEFAULT 0,
    last_login_at       TIMESTAMPTZ,
    extra               JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    CONSTRAINT uq_users_uuid     UNIQUE (uuid),
    CONSTRAINT uq_users_email    UNIQUE (email),
    CONSTRAINT uq_users_username UNIQUE (username)
);
CREATE INDEX idx_users_status_deleted ON users (status, deleted_at);

CREATE TABLE IF NOT EXISTS roles (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(64) NOT NULL,
    description TEXT,
    is_system   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT uq_roles_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS permissions (
    id          BIGSERIAL PRIMARY KEY,
    resource    VARCHAR(64) NOT NULL,
    action      VARCHAR(32) NOT NULL,
    description TEXT,
    CONSTRAINT uq_permissions_resource_action UNIQUE (resource, action)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id    BIGINT      NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS oauth2_clients (
    id                  BIGSERIAL PRIMARY KEY,
    client_id           VARCHAR(64)  NOT NULL,
    client_secret_hash  VARCHAR(255) NOT NULL,
    name                VARCHAR(128) NOT NULL,
    redirect_uris       JSONB        NOT NULL DEFAULT '[]',
    scopes              JSONB        NOT NULL DEFAULT '[]',
    grant_types         JSONB        NOT NULL DEFAULT '["authorization_code"]',
    is_active           BOOLEAN      NOT NULL DEFAULT TRUE,
    owner_user_id       BIGINT       NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    CONSTRAINT uq_oauth2_clients_client_id UNIQUE (client_id)
);

CREATE TABLE IF NOT EXISTS oauth2_authorization_codes (
    id                    BIGSERIAL PRIMARY KEY,
    code                  VARCHAR(128) NOT NULL,
    client_id             VARCHAR(64)  NOT NULL,
    user_id               BIGINT       NOT NULL REFERENCES users(id),
    redirect_uri          TEXT         NOT NULL,
    scopes                JSONB        NOT NULL DEFAULT '[]',
    code_challenge        VARCHAR(128),
    code_challenge_method VARCHAR(8),
    expires_at            TIMESTAMPTZ  NOT NULL,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_auth_codes_code UNIQUE (code)
);
CREATE INDEX idx_auth_codes_code_expires ON oauth2_authorization_codes (code, expires_at);

CREATE TABLE IF NOT EXISTS oauth2_tokens (
    id            BIGSERIAL PRIMARY KEY,
    jti           VARCHAR(64)  NOT NULL,
    client_id     VARCHAR(64)  NOT NULL,
    user_id       BIGINT       NOT NULL REFERENCES users(id),
    scopes        JSONB        NOT NULL DEFAULT '[]',
    refresh_token VARCHAR(512) NOT NULL,
    expires_at    TIMESTAMPTZ  NOT NULL,
    revoked_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_oauth2_tokens_jti           UNIQUE (jti),
    CONSTRAINT uq_oauth2_tokens_refresh_token UNIQUE (refresh_token)
);
CREATE INDEX idx_oauth2_tokens_user_id ON oauth2_tokens (user_id);

INSERT INTO permissions (resource, action, description) VALUES
    ('user','create','创建用户'),('user','read','查看用户'),
    ('user','update','修改用户'),('user','delete','删除用户'),
    ('role','create','创建角色'),('role','read','查看角色'),
    ('role','update','修改角色'),('role','delete','删除角色'),
    ('oauth2_client','create','注册OAuth应用'),('oauth2_client','read','查看OAuth应用')
ON CONFLICT DO NOTHING;

INSERT INTO roles (name, description, is_system) VALUES
    ('admin','超级管理员',TRUE),('editor','编辑者',TRUE),('viewer','只读用户',TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_oauth2_clients_updated_at BEFORE UPDATE ON oauth2_clients FOR EACH ROW EXECUTE FUNCTION set_updated_at();
