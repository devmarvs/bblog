-- +migrate Up
CREATE SCHEMA IF NOT EXISTS bblog;

CREATE TABLE IF NOT EXISTS bblog.app_versions (
    version_id BIGSERIAL PRIMARY KEY,
    api_version VARCHAR NOT NULL,
    mobile_version VARCHAR NOT NULL,
    created_ts TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS app_versions_created_idx ON bblog.app_versions (created_ts DESC, version_id DESC);

-- +migrate Down
DROP INDEX IF EXISTS app_versions_created_idx;
DROP TABLE IF EXISTS bblog.app_versions;
