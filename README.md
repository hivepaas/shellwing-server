# shellwing server

The server of [shellwing](https://github.com/hivepaas/shellwing): it syncs each person's encrypted data between their
devices, and issues short-lived SSH certificates. It never reads what it syncs: encryption keys stay on the
devices.

**Status: early work, not ready to run in production.** The app syncs with it, and signs in to SSH servers with
the certificates it signs.

Anyone may host it. Up to 2 users are free; a license lifts that limit. shellwing cloud is this same server, run by
its maintainers.

## What it needs

- PostgreSQL, for users, settings and a line of metadata for each synced file.
- Redis, for sessions, rate limits, background tasks and notifying other instances.
- A local folder or S3-compatible storage, for the content of synced files.

## Development

```sh
make init                                                                  # builds the tools image
docker stack deploy -c deployment/local/shellwing_server.yaml shellwing_server  # Postgres on 47432, Redis on 47379
make seed-data-with-clear                                                  # WIPES the local database, then fills it
make run                                                                   # on port 10200; the web app, yarn dev, on 4200
```

Tests that need real servers skip without them: `SHW_TEST_PG_DSN`, a PostgreSQL with the migrations applied, and
`SHW_TEST_S3_ENDPOINT`, `_BUCKET`, `_ACCESS_KEY`, `_SECRET_KEY`. `scripts/e2e.sh` runs the app's tests that need a
server against one of its own, on a database it makes and removes; it needs the app checked out next to this one.

A migration in a release (v0.1.0 on) is never edited, since servers have run it: a change is a new migration, and
CI refuses an edit. `make seed-data-with-clear` makes the local database again. The web app is in [web/](web/README.md).

## License

[AGPL-3.0](LICENSE). The web app was made from `hivepaas-dashboard` and keeps its Apache-2.0 [notice](web/LICENSE).
