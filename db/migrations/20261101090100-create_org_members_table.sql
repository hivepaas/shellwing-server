-- +migrate Up
-- Who belongs to which organization, and as what. users.role says what a person may do to the server itself; this says
-- what they may do in one organization, and neither follows from the other. A user is not tied to one: in shellwing
-- cloud a person can be in several.
--
-- status is the organization's say: an admin suspends a member, or removes them. A removed member's line stays, so
-- that saving their account does not make them a member again. Whether the account itself may sign in is users.status.
CREATE TABLE IF NOT EXISTS org_members
(
    org_id     VARCHAR(100) NOT NULL,
    user_id    VARCHAR(100) NOT NULL,
    role       VARCHAR NOT NULL CONSTRAINT chk_role CHECK
                   (role IN ('owner','admin','member')),
    status     VARCHAR NOT NULL CONSTRAINT chk_status CHECK
                   (status IN ('active','suspended','removed')) DEFAULT 'active',
    -- The accounts an SSH certificate of this person may log in as. None until an admin grants them.
    ssh_principals TEXT[] NOT NULL DEFAULT '{}',

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (org_id, user_id),
    CONSTRAINT fk_org_members_org_id FOREIGN KEY (org_id) REFERENCES organizations (id),
    CONSTRAINT fk_org_members_user_id FOREIGN KEY (user_id) REFERENCES users (id)
);

-- org_id is the leftmost column of the primary key; this is for a user's organizations.
CREATE INDEX idx_org_members_user_id ON org_members(user_id);

-- +migrate Down
DROP TABLE IF EXISTS org_members;
