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

## Backups

Back up PostgreSQL. Redis holds nothing that cannot be lost: people sign in again at worst.
