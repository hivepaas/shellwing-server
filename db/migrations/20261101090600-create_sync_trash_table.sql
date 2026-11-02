-- +migrate Up
-- Content in the store that no line may need any more: what a write replaced, a file removed, or a write that lost.
-- It goes after a delay, so that a read already under way still finds it, and only if no line names it by then: the
-- same content can be written back, under the same key.
CREATE TABLE IF NOT EXISTS sync_trash
(
    space_id     VARCHAR(100) NOT NULL,
    name         VARCHAR(255) NOT NULL,
    etag         VARCHAR(64) NOT NULL,
    delete_after TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (space_id, name, etag)
);

CREATE INDEX idx_sync_trash_delete_after ON sync_trash(delete_after);

-- +migrate Down
DROP TABLE IF EXISTS sync_trash;
