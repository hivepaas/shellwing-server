-- +migrate Up
-- An authority that the target servers trust but that does not sign.
--
-- Rotating an authority has two halves, and both need this state. A new one is published before it signs anything, so
-- that every server has picked it up first - signing straight away would refuse everybody whose server had not been
-- updated yet, which is the opposite of what rotating is for. And the one being replaced stays published afterwards,
-- so the certificates it already signed keep working until they expire rather than everybody being refused at once.
--
-- Replacing an authority outright - the key is compromised, nothing it signed should be taken again - is still a thing
-- to do, and still goes straight to 'retired'.
ALTER TABLE ssh_cas
    DROP CONSTRAINT chk_status;
ALTER TABLE ssh_cas
    ADD CONSTRAINT chk_status CHECK (status IN ('active', 'published', 'retired'));

-- A local authority holds its private key until it is retired, not until it stops signing: a published one has not
-- signed yet, or will not sign again, but in both cases it may be asked to.
ALTER TABLE ssh_cas
    DROP CONSTRAINT chk_key_located;
ALTER TABLE ssh_cas
    ADD CONSTRAINT chk_key_located CHECK (
        CASE key_provider
            WHEN 'local' THEN (status <> 'retired') = (private_key IS NOT NULL)
                              AND key_name = '' AND key_version = 0
            ELSE private_key IS NULL AND key_name <> '' AND key_version > 0
        END
    );

-- idx_uq_ssh_cas_active stays as it is: exactly one authority signs.

-- +migrate Down
ALTER TABLE ssh_cas
    DROP CONSTRAINT IF EXISTS chk_key_located;
UPDATE ssh_cas SET status = 'retired', private_key = NULL, retired_at = COALESCE(retired_at, CURRENT_TIMESTAMP)
WHERE status = 'published';
ALTER TABLE ssh_cas
    DROP CONSTRAINT IF EXISTS chk_status;
ALTER TABLE ssh_cas
    ADD CONSTRAINT chk_status CHECK (status IN ('active', 'retired'));
ALTER TABLE ssh_cas
    ADD CONSTRAINT chk_key_located CHECK (
        CASE key_provider
            WHEN 'local' THEN (status = 'active') = (private_key IS NOT NULL)
                              AND key_name = '' AND key_version = 0
            ELSE private_key IS NULL AND key_name <> '' AND key_version > 0
        END
    );
