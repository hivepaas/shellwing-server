-- +migrate Up
-- A collection: a space an organization shares (docs/design/collections.md in the app's repository). Its id is its
-- space's. The name is in the clear, so that admins can manage access on the web; what is in it is not.
--
-- key_version: the version of the key its records are sealed with now. 0 until an app makes the first key: the web,
-- which has none, can make a collection, and the first app with Manage gives it its key.
CREATE TABLE IF NOT EXISTS collections
(
    id          VARCHAR(100) PRIMARY KEY,
    org_id      VARCHAR(100) NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    key_version INT4 NOT NULL DEFAULT 0,
    created_by  VARCHAR(100) NULL,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ NULL,

    CONSTRAINT fk_collections_id FOREIGN KEY (id) REFERENCES spaces (id),
    CONSTRAINT fk_collections_org_id FOREIGN KEY (org_id) REFERENCES organizations (id)
);

CREATE UNIQUE INDEX idx_uq_collections_org_id_name ON collections(org_id, lower(name)) WHERE deleted_at IS NULL;

-- Who may do what in a collection: a person or a group, granted Connect, Edit or Manage. What someone has is the most
-- of their own grant and their groups'. Owners of the organization have Manage on every collection without a line.
CREATE TABLE IF NOT EXISTS collection_grants
(
    id            VARCHAR(100) PRIMARY KEY,
    collection_id VARCHAR(100) NOT NULL,
    org_id        VARCHAR(100) NOT NULL,
    user_id       VARCHAR(100) NULL,
    group_id      VARCHAR(100) NULL,
    permission    VARCHAR NOT NULL CONSTRAINT chk_permission CHECK (permission IN ('connect','edit','manage')),

    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_grantee CHECK ((user_id IS NULL) <> (group_id IS NULL)),
    CONSTRAINT fk_collection_grants_collection_id FOREIGN KEY (collection_id) REFERENCES collections (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_collection_grants_member FOREIGN KEY (org_id, user_id)
        REFERENCES org_members (org_id, user_id) ON DELETE CASCADE,
    CONSTRAINT fk_collection_grants_group_id FOREIGN KEY (group_id) REFERENCES org_groups (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_uq_collection_grants_user ON collection_grants(collection_id, user_id)
    WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX idx_uq_collection_grants_group ON collection_grants(collection_id, group_id)
    WHERE group_id IS NOT NULL;
CREATE INDEX idx_collection_grants_org_id_user_id ON collection_grants(org_id, user_id);

-- A collection's key, wrapped by an app for one person: an ephemeral X25519 key agreed with theirs, the key sealed
-- with what that gives, and a commitment to the key, signed by whoever wrapped it. The server checks the signature and
-- that the key is the person's, never what is sealed. The wrapper's keys are kept as they were, so that the signature
-- still checks once they have another identity.
CREATE TABLE IF NOT EXISTS collection_key_wraps
(
    collection_id    VARCHAR(100) NOT NULL,
    user_id          VARCHAR(100) NOT NULL,
    key_version      INT4 NOT NULL,
    recipient_x25519 BYTEA NOT NULL,
    ephemeral_public BYTEA NOT NULL,
    sealed_key       BYTEA NOT NULL,
    commitment       BYTEA NOT NULL,
    wrapped_by       VARCHAR(100) NOT NULL,
    wrapper_x25519   BYTEA NOT NULL,
    wrapper_ed25519  BYTEA NOT NULL,
    signature        BYTEA NOT NULL,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (collection_id, user_id, key_version),
    CONSTRAINT fk_collection_key_wraps_collection_id FOREIGN KEY (collection_id) REFERENCES collections (id)
        ON DELETE CASCADE,
    CONSTRAINT fk_collection_key_wraps_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_collection_key_wraps_user_id ON collection_key_wraps(user_id);

-- +migrate Down
DROP TABLE IF EXISTS collection_key_wraps;
DROP TABLE IF EXISTS collection_grants;
DROP TABLE IF EXISTS collections;
