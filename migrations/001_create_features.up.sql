CREATE TABLE IF NOT EXISTS features (
    id          BIGSERIAL PRIMARY KEY,
    slug        VARCHAR(128) NOT NULL,
    name        VARCHAR(256) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    status      VARCHAR(16)  NOT NULL DEFAULT 'draft',
    owner_uuid  VARCHAR(36)  NOT NULL,
    owner_email VARCHAR(255) NOT NULL,
    tags        JSONB        NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    CONSTRAINT uq_features_slug UNIQUE (slug),
    CONSTRAINT chk_feature_status CHECK (status IN ('draft','active','archived'))
);
CREATE INDEX idx_features_owner_uuid ON features (owner_uuid);
CREATE INDEX idx_features_status     ON features (status) WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION set_features_updated_at() RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_features_updated_at
    BEFORE UPDATE ON features FOR EACH ROW EXECUTE FUNCTION set_features_updated_at();
