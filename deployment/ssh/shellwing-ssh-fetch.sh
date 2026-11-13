#!/bin/sh
# Fetches the two files sshd needs from shellwing, and puts them where it reads them.
#
#   TrustedUserCAKeys  the authorities whose certificates this server accepts
#   RevokedKeys        the certificates it must refuse before they expire
#
# Both come from the same place with the same token, and they are useless apart: a server with the keys and not the
# list takes certificates that were revoked, and one with the list and not the keys takes none at all.
#
# Keeping the keys up to date is what makes rotating an authority possible. A new one is published before anything is
# signed with it, so a server that runs this on a timer has the new key before any certificate carries it. Without
# that, the moment an organization rotated, everybody whose server had not been edited by hand would be refused.
#
# sshd reads both files on every login, so nothing needs reloading and the only delay is how often this runs.
#
# It reads its settings from /etc/shellwing/ssh.conf:
#
#   SHELLWING_URL=https://shellwing.example.com
#   SHELLWING_ORG=01JC...
#   SHELLWING_TOKEN=...                               # from Administration > SSH Certificates
#   SHELLWING_CA_KEYS_FILE=/etc/ssh/shellwing_ca_keys.pub
#   SHELLWING_KRL_FILE=/etc/ssh/shellwing_krl
#
# Keep that file readable by root only: the token fetches this organization's two files, and nothing else.

set -eu

CONF=${SHELLWING_SSH_CONF:-/etc/shellwing/ssh.conf}
[ -r "$CONF" ] && . "$CONF"

: "${SHELLWING_URL:?set SHELLWING_URL in $CONF}"
: "${SHELLWING_ORG:?set SHELLWING_ORG in $CONF}"
# SHELLWING_KRL_TOKEN is what this used to be called, when it fetched only the list.
TOKEN=${SHELLWING_TOKEN:-${SHELLWING_KRL_TOKEN:-}}
[ -n "$TOKEN" ] || { echo "set SHELLWING_TOKEN in $CONF" >&2; exit 1; }
CA_KEYS_FILE=${SHELLWING_CA_KEYS_FILE:-/etc/ssh/shellwing_ca_keys.pub}
KRL_FILE=${SHELLWING_KRL_FILE:-/etc/ssh/shellwing_krl}

# looks_like_ca_keys: a TrustedUserCAKeys file, whatever kind of key the authorities are.
#
# Every SSH public key blob begins with the length of its type string, four bytes, which base64 encodes to AAAA - so
# this says "the second field is an SSH key" without naming an algorithm. Checking for "ssh-" instead would pass an
# Ed25519 authority and refuse an ECDSA one.
looks_like_ca_keys() {
    awk 'NR == 1 { exit !($2 ~ /^AAAA/) } END { if (NR == 0) exit 1 }' "$1"
}

# looks_like_krl: a revocation list, which OpenSSH's own format marks.
looks_like_krl() {
    [ "$(dd if="$1" bs=1 count=6 2>/dev/null)" = "SSHKRL" ]
}

# fetch <api path> <destination> <the function that says it is one> <what it is>
#
# Written beside the real file and moved into place, so that sshd never reads a half-written one. Every failure leaves
# the file already there alone: sshd told to read a file it cannot parse refuses every certificate, which is the one
# way this script could lock the machine out.
fetch() {
    path=$1
    dest=$2
    looks_like=$3
    what=$4

    tmp="$dest.new.$$"
    # shellcheck disable=SC2064 # tmp is wanted as it is now, not as it may be later
    trap "rm -f '$tmp'" EXIT INT TERM

    if ! curl --fail --silent --show-error --location --max-time 30 \
        --header "X-Shellwing-KRL-Token: $TOKEN" \
        --output "$tmp" \
        "$SHELLWING_URL/api/orgs/$SHELLWING_ORG/$path"
    then
        echo "shellwing-ssh-fetch: could not fetch the $what, keeping the one in place" >&2
        return 1
    fi

    # An empty answer is a failure curl did not call one. An organization with no authority yet has nothing to accept
    # anyway, so keeping what is there is no worse and is safer than installing nothing.
    if [ ! -s "$tmp" ]; then
        echo "shellwing-ssh-fetch: the server returned an empty $what, keeping the one in place" >&2
        return 1
    fi

    # And it reads as that kind of file. A sign-in page or a proxy's error page would otherwise be moved into place,
    # and sshd refusing to read the file refuses every certificate with it.
    if ! "$looks_like" "$tmp"; then
        echo "shellwing-ssh-fetch: the answer is not a $what, keeping the one in place" >&2
        return 1
    fi

    chmod 0644 "$tmp"
    mv -f "$tmp" "$dest"
    trap - EXIT INT TERM
}

# The keys first: a server that has picked up a new authority but not the latest revocations is in a better state than
# one that refuses certificates it should accept.
failed=0
fetch "ssh/ca-keys" "$CA_KEYS_FILE" looks_like_ca_keys "certificate authority keys" || failed=1
fetch "ssh/krl" "$KRL_FILE" looks_like_krl "revocation list" || failed=1
exit "$failed"
