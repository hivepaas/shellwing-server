#!/bin/bash
#
# The app against this server, end to end: a server of its own on a new
# database, then the app's tests that need one (shellwing's
# crates/core/tests/shellwing_server.rs).
#
# Needs the local stack (README: Development) and a checkout of the app, by
# default next to this one: SHELLWING_DIR says where otherwise. The database
# shellwing_e2e is made again at each run and removed after.

set -euo pipefail

cd "$(dirname "$0")/.."
APP_DIR="${SHELLWING_DIR:-../shellwing}"
PORT="${E2E_PORT:-10210}"
DB=shellwing_e2e
WORK="$(mktemp -d)"
devtools() {
    docker run --user "$(id -u)" --rm --volume "$PWD":/app --network=host \
        -e SHW_DB_HOST=localhost -e SHW_DB_PORT=47432 -e SHW_DB_USER=shellwing -e SHW_DB_PASSWORD=abc123 \
        -e SHW_DB_DB_NAME=$DB shellwing-devtools "$@"
}
psql() {
    devtools psql -q -d "host=localhost port=47432 dbname=shellwing user=shellwing password=abc123" "$@"
}

export SHW_ENV=development
export SHW_CONFIG_FILE=config/config.local.toml
export SHW_DB_DB_NAME=$DB
export SHW_HTTP_SERVER_PORT=$PORT
# A Redis database of its own, and a folder for the app's files and the sync store.
export SHW_CACHE_URL="redis://default:abc123@localhost:47379/9"
export SHW_APP_PATH="$WORK/app"
export SHW_USER_ADMIN_EMAIL=e2e@example.test
export SHW_USER_ADMIN_USERNAME=e2e
export SHW_USER_ADMIN_PASSWORD="E2e-$(head -c 12 /dev/urandom | base64 | tr -d '/+=')"

server=""
finish() {
    [ -n "$server" ] && kill "$server" 2>/dev/null && wait "$server" 2>/dev/null
    psql -c "DROP DATABASE IF EXISTS $DB WITH (FORCE)" || true
    rm -rf "$WORK"
}
trap finish EXIT

psql -c "DROP DATABASE IF EXISTS $DB WITH (FORCE)" -c "CREATE DATABASE $DB"
devtools sql-migrate up -config=db/dbconfig.yml -env=main >/dev/null
docker exec "$(docker ps -qf name=shellwing_server_redis)" \
    redis-cli --no-auth-warning -a abc123 -n 9 flushdb >/dev/null

go build -o "$WORK/server" ./cmd/app/...
"$WORK/server" >"$WORK/server.log" 2>&1 &
server=$!
for _ in $(seq 60); do
    curl -s -o /dev/null "http://localhost:$PORT/api/orgs" && break
    if ! kill -0 "$server" 2>/dev/null; then
        cat "$WORK/server.log"
        exit 1
    fi
    sleep 1
done

status=0
(
    cd "$APP_DIR"
    SHELLWING_TEST_SERVER="http://localhost:$PORT" \
        SHELLWING_TEST_SERVER_USER="$SHW_USER_ADMIN_USERNAME" \
        SHELLWING_TEST_SERVER_PASSWORD="$SHW_USER_ADMIN_PASSWORD" \
        cargo test -p shellwing-core --test shellwing_server -- --nocapture
) || status=$?
if [ "$status" -ne 0 ]; then
    echo "--- the server's log (last 50 lines) ---"
    tail -50 "$WORK/server.log"
fi
exit "$status"
