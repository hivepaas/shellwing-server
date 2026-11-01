#!/bin/sh
# The image's entrypoint: brings the database's schema up to date, then becomes
# the server. Run one replica while a new version migrates: a migration that
# fails stops the start, with sql-migrate's error in the logs.
#
# With a command - docker run <image> sql-migrate status -config=db/dbconfig.yml
# -env=main, or sh - the command runs instead, and nothing is migrated.
set -eu

if [ "$#" -gt 0 ]; then
  exec "$@"
fi

fail() {
  echo "shellwing-entrypoint: $*" >&2
  exit 1
}

# sql-migrate reads the database's settings from the environment, through
# db/dbconfig.yml (its environment `main`), not from the server's config file.
for name in SHW_DB_HOST SHW_DB_PORT SHW_DB_USER SHW_DB_PASSWORD SHW_DB_DB_NAME; do
  eval "value=\${$name:-}"
  [ -n "$value" ] || fail "$name is not set: the migrations read the database's settings from the environment"
done

# db/dbconfig.yml puts the password into a URL as it is, so a character a URL
# gives a meaning to would change the address sql-migrate connects to.
case "$SHW_DB_PASSWORD" in
  *[!A-Za-z0-9._~-]*)
    fail "SHW_DB_PASSWORD holds a character a URL must escape:" \
      "use letters and digits, such as a hex one (openssl rand -hex 32)"
    ;;
esac

echo "shellwing-entrypoint: migrating the database"
sql-migrate up -config=db/dbconfig.yml -env=main

exec ./shellwing_server
