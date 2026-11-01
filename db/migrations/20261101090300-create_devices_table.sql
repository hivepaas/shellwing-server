-- +migrate Up
-- A signed-in app. Its refresh token is here rather than in Redis, whose data may be lost: losing it must not sign
-- every device out for good. Only the token's hash is kept.
CREATE TABLE IF NOT EXISTS devices
(
    id                 VARCHAR(100) PRIMARY KEY,
    user_id            VARCHAR(100) NOT NULL,
    name               VARCHAR(100) NOT NULL DEFAULT '',
    platform           VARCHAR(50) NULL,
    refresh_token_hash VARCHAR(100) NULL,

    created_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at       TIMESTAMPTZ NULL,
    refresh_expire_at  TIMESTAMPTZ NULL,
    revoked_at         TIMESTAMPTZ NULL,

    CONSTRAINT fk_devices_user_id FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE UNIQUE INDEX idx_uq_devices_refresh_token_hash ON devices(refresh_token_hash) WHERE refresh_token_hash IS NOT NULL;

-- +migrate Down
DROP TABLE IF EXISTS devices;
