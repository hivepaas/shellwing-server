-- +migrate Up
-- A set of records that sync together and share one key. A personal space belongs to its user and to no organization,
-- which never sees it; a collection belongs to an organization. Names are not here: they are the clients', encrypted.
CREATE TABLE IF NOT EXISTS spaces
(
    id            VARCHAR(100) PRIMARY KEY,
    type          VARCHAR NOT NULL CONSTRAINT chk_type CHECK
                      (type IN ('personal','collection')),
    org_id        VARCHAR(100) NULL,
    owner_user_id VARCHAR(100) NULL,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMPTZ NULL,

    CONSTRAINT chk_owner CHECK (
        (type = 'personal' AND owner_user_id IS NOT NULL AND org_id IS NULL) OR
        (type = 'collection' AND org_id IS NOT NULL AND owner_user_id IS NULL)
    ),
    CONSTRAINT fk_spaces_org_id FOREIGN KEY (org_id) REFERENCES organizations (id),
    CONSTRAINT fk_spaces_owner_user_id FOREIGN KEY (owner_user_id) REFERENCES users (id)
);

CREATE UNIQUE INDEX idx_uq_spaces_owner_user_id ON spaces(owner_user_id) WHERE type = 'personal' AND deleted_at IS NULL;
CREATE INDEX idx_spaces_org_id ON spaces(org_id);

-- +migrate Down
DROP TABLE IF EXISTS spaces;
