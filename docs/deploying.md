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
| `SHW_KMS_PROVIDER` | What holds the key the stored secrets are under: `local` (the default, the app secret), `vault-transit` or `awskms`. See below. |
| `SHW_KMS_VAULT_ADDRESS`, `_TOKEN`, `_KEY` | For `vault-transit`: the Vault or OpenBao, a token of the server's own, and the transit key's name. |
| `SHW_KMS_VAULT_MOUNT`, `_NAMESPACE`, `_CA_CERT_FILE` | Where transit is mounted (`transit`), the Vault Enterprise namespace, and a PEM to trust instead of the system store. |
| `SHW_KMS_AWS_KEY_ID`, `_REGION` | For `awskms`: the key's ID, ARN or alias, and its region. |
| `SHW_KMS_AWS_ACCESS_KEY_ID`, `_SECRET_ACCESS_KEY`, `_SESSION_TOKEN` | Its credential. Leave all three unset to use the instance or service account role, which is better: no long-lived secret in the configuration. |

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
   `PUT /api/orgs/{org}/ssh/members/{user}/principals` with `{"principals": ["deploy"]}`, or gives them to a group
   (*Administration › Groups*, or `/api/orgs/{org}/groups`) whose members all get them. A person's certificates name
   their own accounts and their groups', at most 32 together. Nobody has any until granted, owners included. A username
   is never one by itself: whoever could take the name `root` would get root on every server.

A certificate lasts 2 hours unless `PUT /api/orgs/{org}/ssh/settings` says otherwise, from 5 minutes to a day. Each one
signed is in the audit log (`ssh-cert-issue`), and `sshd` logs the person and device of each login. One person may be
signed 60 an hour, and one of their devices 30: ordinary use is a handful a day, and the limit bounds what a stolen
token can collect.

**Nobody grants themselves.** An admin may grant anybody else, and may take accounts away from themselves, but a change
that would give them an account they did not already have is refused - whether through their own accounts, a group's,
or by joining a group. Another admin grants it, and the audit log says who. Without that rule one admin account taken
over is root on every server the CA is trusted on. It is not a complete separation: an admin who can invite a user can
still invite one whose mailbox is theirs, so grant the admin role as narrowly as the accounts themselves.

**What a certificate may do.** A certificate carries the extensions `sshd` checks, and anything it does not carry is
refused however the target server is configured - so this is the narrower of the two gates. By default a certificate
allows a terminal and nothing else: `permit-pty`. Port forwarding is off, because it is how somebody reaches a machine
they were never granted through one they were, and so is agent forwarding, which hands the person's keys to whoever
has root on the server they logged in to. `ssh-keygen -s` grants all five by default, which suits a person signing one
certificate for themselves rather than a server signing them for everybody.

*Administration › SSH Certificates* chooses them, or `PUT /api/orgs/{org}/ssh/settings` with
`{"certExtensions": ["permit-pty", "permit-port-forwarding"]}`. Both fields of that call are optional and what is left
out is left alone. An empty list is a setting of its own: a certificate that logs in and runs a command, with no
terminal. Changing them revokes the certificates people hold, since each of those still allows what it allowed when it
was signed.

**Taking access away.** A change to what somebody may log in as revokes the certificates they already hold, so that
what they have matches what they are granted. These all revoke: their own accounts changed or taken away, a group of
theirs changing what it grants or being deleted, being taken out of a group, being suspended or removed from the
organization, their account being disabled or deleted, the device holding it being revoked, and a change to what
certificates may do. The app asks for a new
certificate when it next needs one, so somebody who still has accounts carries on; somebody who does not, stops.

Revoking puts the certificate's serial in the organization's revocation list, which the target servers read through
`RevokedKeys` - see below. Until a server has fetched the list, what was revoked still works there until it ends,
which is the other reason to keep the lifetime short.

**Replacing the authority.** Making a new one retires the old at once, destroys its private key, and revokes
everything it signed: from then on certificates are signed by the new one only. Add its public key to the target
servers beside the old one straight away, and remove the old one when its last certificates have ended. If the key may
have leaked, remove the old one first.

**The private key** is in PostgreSQL, sealed with the server's data key like any stored secret. Anybody with both the
database and the app secret can sign for every account the target servers let a certificate reach, so treat a backup of
them as you would the target servers' root.

### Hardening the target servers

What the CA says is only half of it: the other half is each server deciding what it will accept. These cost nothing to
set up and they are what keeps one compromised CA from being root everywhere.

**Let each server name the accounts it accepts.** `TrustedUserCAKeys` on its own means any principal the CA signs is
good on that server. With an `AuthorizedPrincipalsFile`, the server decides too, and a certificate naming `root` is
worth nothing on a server whose file does not list it:

```
# /etc/ssh/sshd_config
TrustedUserCAKeys /etc/ssh/shellwing_ca.pub
AuthorizedPrincipalsFile /etc/ssh/principals/%u
```

`/etc/ssh/principals/deploy` then holds the principals that may become the `deploy` account, one per line. Keep these
files with the server's configuration, not from this server: the point is that the two have to agree.

**Read the revocation list.** This is what makes taking access away take effect before a certificate ends. The server
keeps a list per organization and `sshd` refuses everything in it:

```
# /etc/ssh/sshd_config
RevokedKeys /etc/ssh/shellwing_krl
```

`GET /api/orgs/{org}/ssh/krl` returns that list, the binary format `ssh-keygen -k` writes and `sshd` reads. It covers
every authority the organization has had, so one file is enough however many times the CA has been replaced, and it is
a valid list even with nothing revoked yet.

**Giving a server a token.** A server has no account here, so it fetches with a token of its own.
*Administration › SSH Certificates* makes one, or `POST /api/orgs/{org}/ssh/krl/token`; it is in that one answer and
nowhere else, since only its hash is kept. Making another replaces it, which is how one that may have leaked is
retired, and `DELETE /api/orgs/{org}/ssh/krl/token` takes it away.

The token fetches one organization's revocation list and does nothing else. That is the point of it: the alternative
is a person's session token on every target server, which can do everything that person can. A server that is taken
over gives up a list of random serial numbers and the public keys already sitting in its own `/etc/ssh`.

A member may also fetch the list with their own session, as they may fetch the CA's public key - handy for a look, and
for a configuration-management run that already holds an operator's credentials.

**Running the fetch.** [`deployment/krl/`](../deployment/krl/) holds a script and a systemd timer that do it:

```sh
install -m 0755 shellwing-krl-fetch.sh /usr/local/bin/
install -m 0644 shellwing-krl.service shellwing-krl.timer /etc/systemd/system/
install -d -m 0700 /etc/shellwing
cat > /etc/shellwing/krl.conf <<'CONF'
SHELLWING_URL=https://shellwing.example.com
SHELLWING_ORG=01JC...
SHELLWING_KRL_TOKEN=...
CONF
chmod 0600 /etc/shellwing/krl.conf
systemctl enable --now shellwing-krl.timer
```

It runs every minute, so a revocation reaches a server within about that. `sshd` reads the file on each login, so
nothing is reloaded. The script writes to a temporary name, checks the answer really is a revocation list, and only
then moves it into place - a server told to read a file it cannot parse refuses every certificate, so a half-written
file or a proxy's error page would lock everybody out of that machine.

**Nothing is pushed.** The server cannot reach your target servers and does not know they exist: hosts live in each
person's encrypted records, which the server cannot read, and it holds no credential for any of them. Giving it one
would mean a server that can write to every machine in the fleet, which is a larger thing to lose than the CA key.
So a revocation is as fast as the fetch, and the lifetime bounds the worst case either way.

`ssh-keygen -Qf /etc/ssh/shellwing_krl cert.pub` says what the file in place has to say about one certificate, which
is the way to check a server is really reading what you think it is.

**One authority per environment.** A certificate is good on every server that trusts the CA that signed it, so
production and staging trusting the same key is one blast radius, not two. `TrustedUserCAKeys` takes a file of several
keys, so a server can trust more than one authority - but an organization here has a single active CA, so separate
environments mean separate organizations until it can hold more.

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
alone, while losing the key its own secrets are under would lose everybody's data.

The encryption stops the server from reading or changing a record, not from holding back or deleting one. A device
keeps what it has: a record replayed from before is older than its own, and loses when they merge. A new device could
start from an old copy.

## Collections

A collection is hosts, secrets and snippets an organization shares, synced like a personal space and encrypted the same
way: with a key the server never has. Owners and admins make them, in *Administration › Collections* or with
`/api/orgs/{org}/collections`, and grant people or groups Connect (see the hosts and connect), Edit (change them) or
Manage (grant others, replace the key). Owners have Manage on every collection. An admin manages a collection's access
and reads it only if granted.

**The key.** A collection made on the web has none until the app of somebody with Manage opens it and makes one. People
granted access then wait for it: the app of somebody holding it shows each person's fingerprint (theirs is on their
*Profile*) and, once compared, wraps the key for that person's identity. The server checks each wrap is signed by the
person giving it and made for the right identity, keeps one per person and key version, and serves it only to them.
Somebody holding the key who loses access, by a grant removed, leaving a group or being removed, marks the collection
as needing a new key, which the app of somebody with Manage makes. Suspending a member stops the server serving them,
and is no reason for a new key.

**What the server sees** of a collection: its name and description, who may do what, the size of its records, and how
many are hosts. Each grant, new key, key given and key fetched by a device is in the audit log.

Deleting a collection stops its sync at once; its files go after the store's trash delay.

## Protecting the stored secrets

Stored secrets are encrypted with a data encryption key, and that key is kept in the database wrapped by a key
encryption key. Changing what holds the key encryption key is therefore one row and nothing else: no pass over the
values, nothing to resume, no window where some are readable and some are not.

By default the app secret holds it, which needs nothing deployed beside the server. `SHW_KMS_PROVIDER` moves it to a
service instead:

```
SHW_KMS_PROVIDER=vault-transit
SHW_KMS_VAULT_ADDRESS=https://vault.internal:8200
SHW_KMS_VAULT_TOKEN=...
SHW_KMS_VAULT_KEY=shellwing-dek
```

The token needs three capabilities, and `read` on the key is one of them:

```hcl
path "transit/keys/shellwing-dek"    { capabilities = ["read"] }
path "transit/encrypt/shellwing-dek" { capabilities = ["update"] }
path "transit/decrypt/shellwing-dek" { capabilities = ["update"] }
```

`read` is there because transit's encrypt endpoint **creates** a key it does not know rather than refusing. Without
looking the key up first, a mistyped `SHW_KMS_VAULT_KEY` would quietly get a brand new key, and the key holding every
stored secret would not be the one you named - with none of the rotation, policy or backup you had arranged for that
one, and nothing later to notice, because a new key unwraps its own ciphertext perfectly well. The server looks before
it wraps, and refuses a key that is not there.

What that buys is that the database and its backups stop being enough to read what is in them. Whoever takes both
still has to call the service as this server, which is a separate credential with its own audit trail - and revoking
it retires every copy of the backup at once, which no amount of deleting files does.

What it does not buy is anything against somebody already running as this server. The data encryption key is in its
memory while it runs, and so is an organization's SSH CA key for as long as it signs with it. Keeping the CA key out
of the server altogether is a different change, not this one.

The move happens on the next start, and the log says so. Two things to know about it:

- **Keep the old provider's settings until the move has happened.** Turning a service off is not only a change of
  `SHW_KMS_PROVIDER`: the key still has to be opened by whatever wrapped it. Dropping the service's address and token
  in the same change leaves a server that will not start, and says so. Change the provider first, let it start, then
  remove them.
- **A key the server cannot open stops the start.** That is deliberate. A server that carried on would write new
  secrets under a new key beside the ones it can no longer read.

`vault-transit` works against OpenBao as well, and is the one to choose for a self-hosted install. Of the three, it is
also the only one whose keys can sign Ed25519, which matters if signing itself moves to the service later.

## Backups

Back up PostgreSQL and the sync store together, database first: content a line needs is written before the line, so
a store copied after the database has everything the database names. What the store holds beyond that is removed by
the server in time. Redis holds nothing that cannot be lost.
