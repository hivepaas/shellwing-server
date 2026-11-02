-- +migrate Up
-- This server's identity to the portal, and the license key installed on it. One row. The key is signed and holds
-- nothing secret enough to encrypt: whoever reads it sees who it was issued to, as the License page shows.
CREATE TABLE IF NOT EXISTS license_state
(
    id              INT2 PRIMARY KEY DEFAULT 1 CONSTRAINT chk_one_row CHECK (id = 1),
    installation_id VARCHAR(25) NOT NULL,
    key_text        TEXT NULL,

    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE IF EXISTS license_state;
