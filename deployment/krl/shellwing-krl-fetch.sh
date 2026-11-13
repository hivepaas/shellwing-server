#!/bin/sh
# Fetches the organization's SSH key revocation list and puts it where sshd reads it.
#
# A certificate cannot be taken back once signed, so this is how a revocation reaches this server. sshd reads the file
# on every login, so nothing needs reloading and the only delay is how often this runs.
#
# It reads its settings from /etc/shellwing/krl.conf:
#
#   SHELLWING_URL=https://shellwing.example.com
#   SHELLWING_ORG=01JC...
#   SHELLWING_KRL_TOKEN=...            # from Administration > SSH Certificates
#   SHELLWING_KRL_FILE=/etc/ssh/shellwing_krl
#
# Keep that file readable by root only: the token fetches this organization's list, and nothing else.

set -eu

CONF=${SHELLWING_KRL_CONF:-/etc/shellwing/krl.conf}
[ -r "$CONF" ] && . "$CONF"

: "${SHELLWING_URL:?set SHELLWING_URL in $CONF}"
: "${SHELLWING_ORG:?set SHELLWING_ORG in $CONF}"
: "${SHELLWING_KRL_TOKEN:?set SHELLWING_KRL_TOKEN in $CONF}"
KRL_FILE=${SHELLWING_KRL_FILE:-/etc/ssh/shellwing_krl}

# Written beside the real file and moved into place, so that sshd never reads a half-written one. A server told to
# read a file it cannot parse refuses every certificate, which would lock everybody out of this machine.
TMP="$KRL_FILE.new.$$"
trap 'rm -f "$TMP"' EXIT INT TERM

curl --fail --silent --show-error --location --max-time 30 \
    --header "X-Shellwing-KRL-Token: $SHELLWING_KRL_TOKEN" \
    --output "$TMP" \
    "$SHELLWING_URL/api/orgs/$SHELLWING_ORG/ssh/krl"

# A list always has a header, so an empty answer is a failure that curl did not call one.
if [ ! -s "$TMP" ]; then
    echo "shellwing-krl-fetch: the server returned an empty list, keeping the one in place" >&2
    exit 1
fi

# And it starts with what OpenSSH's own revocation lists start with. A sign-in page or a proxy's error page would
# otherwise be moved into place, and sshd refusing to read the file refuses every certificate with it - the one way
# this script could lock the machine out.
if [ "$(dd if="$TMP" bs=1 count=6 2>/dev/null)" != "SSHKRL" ]; then
    echo "shellwing-krl-fetch: the answer is not a revocation list, keeping the one in place" >&2
    exit 1
fi

chmod 0644 "$TMP"
mv -f "$TMP" "$KRL_FILE"
trap - EXIT INT TERM
