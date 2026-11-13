-- +migrate Up
-- A credential for a target server to fetch the organization's revocation list with.
--
-- It exists because the alternative is worse: the list has to be fetched by the servers themselves, and the only
-- credential they would otherwise have is a person's session token, which can do everything that person can. This one
-- reads one organization's revocation list and nothing else, so a server that is taken over gives up a list of random
-- serial numbers and the public keys already sitting in its own /etc/ssh.
--
-- Only the hash is kept, as a device's refresh token is: the token is shown once, when it is made.
ALTER TABLE organizations
    ADD COLUMN ssh_krl_token_hash VARCHAR(100) NULL,
    ADD COLUMN ssh_krl_token_created_at TIMESTAMPTZ NULL;

-- The token names the organization by itself, so it has to be unique across all of them.
CREATE UNIQUE INDEX idx_uq_organizations_ssh_krl_token_hash ON organizations(ssh_krl_token_hash)
    WHERE ssh_krl_token_hash IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_uq_organizations_ssh_krl_token_hash;
ALTER TABLE organizations
    DROP COLUMN IF EXISTS ssh_krl_token_hash,
    DROP COLUMN IF EXISTS ssh_krl_token_created_at;
