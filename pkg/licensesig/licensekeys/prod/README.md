# Production license public keys

The public halves of the key pair the portal signs license keys with: one ed25519 and one ML-DSA-65 key, as
`<key-id>.pub.pem`. Every product the portal sells shares this pair, so these files are copies of the portal's
`pkg/licensesig/prodkeys/shared/`, byte for byte. What keeps another product's key from passing here is the signing
context, `shellwing-server-license-v1`.

When the pair is replaced, add the new files beside these and release before the portal signs with it. Remove these
only when no key they signed is in force. The steps are under Rotation in the portal's `docs/licensing.md`.

Only public keys go here. The private halves stay with whoever runs the portal, outside any repository.
