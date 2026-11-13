-- +migrate Up
-- A certificate cannot be revoked, only waited out, so the time one is good for is the time a stolen one is good for.
-- Eight hours was a working day; two is enough for a session and far less to wait out. An organization that set its
-- own keeps it: this moves the default only, for the ones that never did.
ALTER TABLE organizations ALTER COLUMN ssh_cert_ttl_minutes SET DEFAULT 120;

-- +migrate Down
ALTER TABLE organizations ALTER COLUMN ssh_cert_ttl_minutes SET DEFAULT 480;
