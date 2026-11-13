-- +migrate Up
-- A certificate that was signed. The certificate itself is not kept - the person holds it, and it is not a secret the
-- server needs - but what it says is, so that one can be named in a revocation list afterwards. Without this row a
-- certificate can only be waited out.
--
-- The serial is the primary key rather than (ca_id, serial), which is all a revocation list needs: a serial is drawn
-- at random and must mean one certificate, and a key that says so is also the key that refuses a collision instead of
-- quietly letting two certificates share a name.
CREATE TABLE IF NOT EXISTS ssh_certificates
(
    -- Within 63 bits, as pkg/sshca draws it, so that it fits a signed BIGINT here and a JSON reader elsewhere.
    serial        BIGINT PRIMARY KEY CONSTRAINT chk_serial CHECK (serial > 0),
    org_id        VARCHAR(100) NOT NULL,
    ca_id         VARCHAR(100) NOT NULL,
    user_id       VARCHAR(100) NOT NULL,
    -- The app that asked, empty for a person on the web app.
    device_id     VARCHAR(100) NOT NULL DEFAULT '',
    -- Of the public key that was signed, as `ssh-keygen -l` writes it.
    fingerprint   VARCHAR(100) NOT NULL DEFAULT '',
    key_id        VARCHAR(255) NOT NULL DEFAULT '',
    principals    TEXT[] NOT NULL DEFAULT '{}',

    valid_after   TIMESTAMPTZ NOT NULL,
    valid_before  TIMESTAMPTZ NOT NULL,
    -- When it was refused, if it was. A row past valid_before needs no revoking: sshd refuses it on its own.
    revoked_at    TIMESTAMPTZ NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_ssh_certificates_org_id FOREIGN KEY (org_id) REFERENCES organizations (id)
);

-- Revoking what one person holds, which is what a grant taken away or a membership ended comes down to.
CREATE INDEX idx_ssh_certificates_org_id_user_id ON ssh_certificates(org_id, user_id);
-- Revoking what one device holds, wherever it holds it: a device revoked is revoked in every organization.
CREATE INDEX idx_ssh_certificates_device_id ON ssh_certificates(device_id) WHERE device_id <> '';
-- Building an organization's revocation list: only what is revoked and has not ended yet goes in it.
CREATE INDEX idx_ssh_certificates_revoked ON ssh_certificates(org_id, valid_before) WHERE revoked_at IS NOT NULL;
-- Clearing out the ones that ended.
CREATE INDEX idx_ssh_certificates_valid_before ON ssh_certificates(valid_before);

-- +migrate Down
DROP TABLE IF EXISTS ssh_certificates;
