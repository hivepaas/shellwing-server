-- +migrate Up
-- A tenant. A self-hosted server has exactly one, made when it is installed and not shown as such; shellwing cloud has
-- one per customer. Whatever belongs to a tenant names it, so a server holding several needs no change to the schema.
CREATE TABLE IF NOT EXISTS organizations
(
    id         VARCHAR(100) PRIMARY KEY,
    slug       VARCHAR(100) NOT NULL,
    name       VARCHAR(255) NOT NULL,
    status     VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                   (status IN ('active','disabled')) DEFAULT 'active',
    update_ver INT4 NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX idx_uq_organizations_slug ON organizations(slug) WHERE deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS organizations;
