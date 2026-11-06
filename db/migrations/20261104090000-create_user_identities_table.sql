-- +migrate Up
-- Each person's identity as their app publishes it: the public halves, which collection keys are wrapped for and which
-- check what they wrap, and the X25519 key signed by the Ed25519 one. The secrets never leave the person's devices.
CREATE TABLE IF NOT EXISTS user_identities
(
    user_id        VARCHAR(100) PRIMARY KEY,
    x25519_public  BYTEA NOT NULL,
    ed25519_public BYTEA NOT NULL,
    signature      BYTEA NOT NULL,
    -- What people compare, worked out from the two keys; kept to show it.
    fingerprint    VARCHAR(40) NOT NULL,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_user_identities_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE IF EXISTS user_identities;
