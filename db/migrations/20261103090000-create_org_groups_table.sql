-- +migrate Up
-- People of an organization bundled to be granted together: a group's SSH accounts are each of its members', on top of
-- their own. A name is unique in its organization, whatever its case.
CREATE TABLE IF NOT EXISTS org_groups
(
    id             VARCHAR(100) PRIMARY KEY,
    org_id         VARCHAR(100) NOT NULL,
    name           VARCHAR(100) NOT NULL,
    description    VARCHAR(500) NOT NULL DEFAULT '',
    -- The accounts an SSH certificate of each member may log in as.
    ssh_principals TEXT[] NOT NULL DEFAULT '{}',

    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_org_groups_org_id FOREIGN KEY (org_id) REFERENCES organizations (id)
);

CREATE UNIQUE INDEX idx_uq_org_groups_org_id_name ON org_groups(org_id, lower(name));

-- Who is in which group. A member of the group is a member of its organization: the line goes with their membership,
-- and a removed member is taken out of every group.
CREATE TABLE IF NOT EXISTS org_group_members
(
    group_id   VARCHAR(100) NOT NULL,
    org_id     VARCHAR(100) NOT NULL,
    user_id    VARCHAR(100) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (group_id, user_id),
    CONSTRAINT fk_org_group_members_group_id FOREIGN KEY (group_id) REFERENCES org_groups (id) ON DELETE CASCADE,
    CONSTRAINT fk_org_group_members_member FOREIGN KEY (org_id, user_id)
        REFERENCES org_members (org_id, user_id) ON DELETE CASCADE
);

-- group_id is the leftmost column of the primary key; this is for a member's groups.
CREATE INDEX idx_org_group_members_org_id_user_id ON org_group_members(org_id, user_id);

-- +migrate Down
DROP TABLE IF EXISTS org_group_members;
DROP TABLE IF EXISTS org_groups;
