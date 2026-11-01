-- +migrate Up
-- One line for each synced file: never its content, which is in the store (a local folder or S3), under a key made of
-- space_id, name and etag. A write is decided here, by updating the line only if etag is still what the writer saw, so
-- conditional writes work the same on every store.
--
-- tombstone: a record deleted by a client, whose file stays until every device has seen it. It is no longer counted.
-- deleted_at: the file itself is gone. The line stays, so that the changes since a revision can say so.
CREATE TABLE IF NOT EXISTS sync_objects
(
    space_id   VARCHAR(100) NOT NULL,
    name       VARCHAR(255) NOT NULL,
    etag       VARCHAR(64) NOT NULL,
    size       BIGINT NOT NULL,
    kind       VARCHAR(20) NULL,
    tombstone  BOOL NOT NULL DEFAULT FALSE,
    revision   BIGINT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,

    PRIMARY KEY (space_id, name),
    CONSTRAINT fk_sync_objects_space_id FOREIGN KEY (space_id) REFERENCES sync_spaces (space_id)
);

CREATE INDEX idx_sync_objects_space_id_revision ON sync_objects(space_id, revision);

-- +migrate Down
DROP TABLE IF EXISTS sync_objects;
