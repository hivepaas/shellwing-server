# Deploying

The server is not ready for production yet: this page lists what a deployment needs today, and grows with it.

## The image

A tag `vX.Y.Z` builds `ghcr.io/hivepaas/shellwing-server:X.Y.Z` (`.github/workflows/release.yml`). At start it brings
the database's schema up to date, then serves the API and the web app on port 10100. Run one replica while a new
version migrates. Give it a volume at `/var/lib/shellwing`.

## Settings

`config/config.production.toml` is the image's configuration. Set these as environment variables:

| Variable | What it is |
| --- | --- |
| `SHW_ROOT_DOMAIN`, `SHW_APP_DOMAIN` | Where the server is reached. |
| `SHW_APP_SECRET` | 64 hex characters. Every stored secret is encrypted with it: back it up elsewhere. |
| `SHW_SESSION_JWT_SECRET` | 64 hex characters. |
| `SHW_DB_HOST`, `SHW_DB_PORT`, `SHW_DB_USER`, `SHW_DB_PASSWORD`, `SHW_DB_DB_NAME` | PostgreSQL. The password in hex: `sql-migrate` puts it in a URL. |
| `SHW_CACHE_URL` | Redis, as `redis://default:<password>@<host>:6379/0`. |
| `SHW_HTTP_SERVER_TRUSTED_PROXIES` | The proxy in front of the server. Rate limits count by the client's address, which only a trusted proxy gives. |
| `SHW_USER_ADMIN_EMAIL`, `SHW_USER_ADMIN_USERNAME`, `SHW_USER_ADMIN_PASSWORD` | The first admin, on the first boot only. Remove them once the admin exists. |
| `SHW_SYNC_STORE_TYPE` | Where synced files are kept: `local` (the default) or `s3`. |
| `SHW_SYNC_STORE_PATH` | The folder of the local store. By default, `sync` in the app path. One instance only. |
| `SHW_SYNC_STORE_ENDPOINT`, `_REGION`, `_BUCKET`, `_PREFIX`, `_PATH_STYLE` | S3-compatible storage, whose bucket must exist. `_PATH_STYLE` is for servers without bucket subdomains. |
| `SHW_SYNC_STORE_ACCESS_KEY`, `SHW_SYNC_STORE_SECRET_KEY` | Its key. It needs to read, write, delete and list. |
| `SHW_SYNC_MAX_OBJECT_BYTES`, `SHW_SYNC_MAX_SPACE_BYTES` | The largest file (8 MiB), and what one space may hold (0: no limit). |

A store that cannot be reached stops the start.

## Users and the license

Up to 2 users are free. A user holds a seat while active or invited; a disabled or deleted one does not. Past the limit
no new user can be made or let back in, and that is all: nobody is removed or locked out, and every feature keeps
working, also when a license ends.

A license key lifts the limit. *Administration › License* in the web app shows this server's ID, which a key is issued
for, and the seats taken, and installs or removes a key; so do `GET`, `PUT` and `DELETE /api/system/license`. A key is verified offline: a server never needs the portal to
run. After a key expires its limit holds for its grace, then the free limit is back.

## SSH certificates

The server can sign short-lived SSH certificates, so that nobody shares a private key and removing a person means they
get no new ones. An organization has none of this until an admin sets it up, in *Administration › SSH Certificates*
in the web app, or with the API below.

1. An admin makes the certificate authority: `POST /api/orgs/{org}/ssh/cas`. The organization's ID is in `GET /api/orgs`.
2. Each target server trusts its public key, with the `publicKey` of `GET /api/orgs/{org}/ssh/cas`:

   ```
   # /etc/ssh/sshd_config
   TrustedUserCAKeys /etc/ssh/shellwing_ca.pub
   ```

3. An admin grants each person the accounts they may log in as:
   `PUT /api/orgs/{org}/ssh/members/{user}/principals` with `{"principals": ["deploy"]}`. Nobody has any until granted,
   owners included. A username is never one by itself: whoever could take the name `root` would get root on every server.

A certificate lasts 8 hours unless `PUT /api/orgs/{org}/ssh/settings` says otherwise, from 5 minutes to a day. Each one
signed is in the audit log (`ssh-cert-issue`), and `sshd` logs the person and device of each login.

**Taking access away.** A certificate cannot be revoked. Removing a person's accounts, or disabling them, stops new ones;
revoking a device stops new ones on that device. The ones they hold work until they end, so keep the lifetime short.

**Replacing the authority.** Making a new one retires the old at once, and destroys its private key: from then on
certificates are signed by the new one only. Add its public key to the target servers beside the old one straight away,
and remove the old one when its last certificates have ended. If the key may have leaked, remove the old one first.

**The private key** is in PostgreSQL, sealed with the server's data key like any stored secret. Anybody with both the
database and the app secret can sign for every account the target servers let a certificate reach, so treat a backup of
them as you would the target servers' root.

## Synced data

The app encrypts each record (a host, a password, a key, a snippet) on the device, with a key the server never has, and
the server stores the file as it comes. Nobody reading the database, the store or a backup of them can read a record,
the server's admins included.

What the server does see, for each person: how many records they have and how many of them are hosts (the app leaves
that in the clear, for plans that count hosts), each file's size, and when each device writes. The records' names are random IDs.

`format.json` in each space holds the vault's key, sealed by the person's recovery key, which cannot be guessed. If
they let a passphrase add devices, it is also sealed by that passphrase: whoever reads the file can try to guess it
offline. The app asks for a passphrase hard to guess and stretches it with Argon2id; keep the store private all the
same, and turn on the encryption it offers (SSE on S3, an encrypted disk for a local folder).

The server does not encrypt the files once more: with its own key, it would only protect against the store leaking
alone, while losing `SHW_APP_SECRET` would lose everybody's data.

The encryption stops the server from reading or changing a record, not from holding back or deleting one. A device
keeps what it has: a record replayed from before is older than its own, and loses when they merge. A new device could
start from an old copy.

Back up PostgreSQL and the sync store together, database first: content a line needs is written before the line, so
a store copied after the database has everything the database names. What the store holds beyond that is removed by
the server in time. Redis holds nothing that cannot be lost.
