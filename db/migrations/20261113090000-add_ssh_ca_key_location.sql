-- +migrate Up
-- Where an authority's private half is, so that it does not have to be here.
--
-- 'local' is what every authority made so far is: an Ed25519 private key in private_key, sealed with the server's data
-- key. 'vault-transit' is a key in a Vault or OpenBao transit engine, named by key_name at key_version, which this
-- server signs with and can never read - the engine makes such a key unexportable.
--
-- key_version is part of the authority, not a detail. A transit key can be rotated, and a rotated key signs under a new
-- public half from then on; an authority whose public key moved would stop being the one every target server was told
-- to trust. So the version is pinned here, and rotating in the engine does nothing to what was already signed.
ALTER TABLE ssh_cas
    ADD COLUMN key_provider VARCHAR NOT NULL DEFAULT 'local'
        CONSTRAINT chk_key_provider CHECK (key_provider IN ('local', 'vault-transit')),
    ADD COLUMN key_name     TEXT NOT NULL DEFAULT '',
    ADD COLUMN key_version  INTEGER NOT NULL DEFAULT 0;

-- The old rule was that an active authority has a private key and a retired one does not. That is still true of a local
-- one, but an authority whose key is in the engine has no private key here at all, active or not - so the rule becomes
-- one per provider, and each says both what must be there and what must not.
ALTER TABLE ssh_cas
    DROP CONSTRAINT chk_private_key;
ALTER TABLE ssh_cas
    ADD CONSTRAINT chk_key_located CHECK (
        CASE key_provider
            WHEN 'local' THEN (status = 'active') = (private_key IS NOT NULL)
                              AND key_name = '' AND key_version = 0
            ELSE private_key IS NULL AND key_name <> '' AND key_version > 0
        END
    );

-- +migrate Down
ALTER TABLE ssh_cas
    DROP CONSTRAINT IF EXISTS chk_key_located;
ALTER TABLE ssh_cas
    DROP COLUMN IF EXISTS key_provider,
    DROP COLUMN IF EXISTS key_name,
    DROP COLUMN IF EXISTS key_version;
ALTER TABLE ssh_cas
    ADD CONSTRAINT chk_private_key CHECK ((status = 'active') = (private_key IS NOT NULL));
