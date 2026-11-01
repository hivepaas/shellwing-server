-- +migrate Up
-- A space's totals, in a row of their own since every write changes them. They can be rebuilt from the store. revision
-- answers "anything new?" without touching the store, and the counts let a limit be checked on each write.
CREATE TABLE IF NOT EXISTS sync_spaces
(
    space_id     VARCHAR(100) PRIMARY KEY,
    revision     BIGINT NOT NULL DEFAULT 0,
    bytes_used   BIGINT NOT NULL DEFAULT 0,
    record_count INT4 NOT NULL DEFAULT 0,
    host_count   INT4 NOT NULL DEFAULT 0,

    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_sync_spaces_space_id FOREIGN KEY (space_id) REFERENCES spaces (id)
);

-- +migrate Down
DROP TABLE IF EXISTS sync_spaces;
