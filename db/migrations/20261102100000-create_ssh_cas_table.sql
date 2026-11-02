-- +migrate Up
-- An organization's SSH certificate authorities. The newest is the active one, the only one that signs. One replaced is
-- retired, with its private key gone: it stays listed so that servers trusting it can be moved over, until it is deleted.
CREATE TABLE IF NOT EXISTS ssh_cas
(
    id          VARCHAR(100) PRIMARY KEY,
    org_id      VARCHAR(100) NOT NULL,
    status      VARCHAR NOT NULL CONSTRAINT chk_status CHECK (status IN ('active','retired')),
    public_key  TEXT NOT NULL,
    private_key TEXT NULL,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    retired_at  TIMESTAMPTZ NULL,

    CONSTRAINT fk_ssh_cas_org_id FOREIGN KEY (org_id) REFERENCES organizations (id),
    CONSTRAINT chk_private_key CHECK ((status = 'active') = (private_key IS NOT NULL))
);

CREATE UNIQUE INDEX idx_uq_ssh_cas_active ON ssh_cas(org_id) WHERE status = 'active';
CREATE INDEX idx_ssh_cas_org_id ON ssh_cas(org_id);

-- +migrate Down
DROP TABLE IF EXISTS ssh_cas;
