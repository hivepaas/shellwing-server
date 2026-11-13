-- +migrate Up
-- What a certificate this organization signs is allowed to do, as sshd's extension names.
--
-- An extension absent from a certificate is refused however sshd is configured, so this is the narrower of the two
-- gates and the one worth setting deliberately. The default is a terminal and nothing else: port forwarding on every
-- certificate is a way to reach a machine nobody was granted, through one they were, and agent forwarding hands the
-- person's keys to whoever has root on the server they logged in to.
--
-- `ssh-keygen -s` grants all five by default, which suits a person signing one certificate for themselves and not a
-- server signing them for everybody.
ALTER TABLE organizations
    ADD COLUMN ssh_cert_extensions TEXT[] NOT NULL DEFAULT '{permit-pty}';

-- +migrate Down
ALTER TABLE organizations
    DROP COLUMN IF EXISTS ssh_cert_extensions;
