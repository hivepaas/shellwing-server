# shellwing server

The server of [shellwing](https://github.com/hivepaas/shellwing): it syncs each person's encrypted data between their
devices, and (planned) issues short-lived SSH certificates. It never reads what it syncs: encryption keys stay on the
devices.

**Status: early work, not ready to run in production.** The sync API works; the app does not use it yet, and the
certificates are still to come.

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
`SHW_TEST_S3_ENDPOINT`, `_BUCKET`, `_ACCESS_KEY`, `_SECRET_KEY`.

While nothing is released, migrations are edited in place rather than added to; run `make seed-data-with-clear` after
each change. The web app is in [web/](web/README.md).

## License

[AGPL-3.0](LICENSE). The web app was made from `hivepaas-dashboard` and keeps its Apache-2.0 [notice](web/LICENSE).
