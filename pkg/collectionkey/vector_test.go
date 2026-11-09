package collectionkey

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

// The vectors of shellwing's docs/design/collections-format.md, made again here from their inputs with Go's own
// libraries: the server only checks wraps, but a second implementation of every step is what tells the document is
// right.

const vectorCollection = "01JC0LLECT10N0000000000000"

func b64(t *testing.T, s string) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(s)
	require.NoError(t, err)
	return b
}

func repeated(b byte, n int) []byte { return bytes.Repeat([]byte{b}, n) }

// bound is the collection id and version as the derivation and the signature bind them.
func bound(collection string, version uint32) []byte {
	out := binary.BigEndian.AppendUint16(nil, uint16(len(collection)))
	out = append(out, collection...)
	return binary.BigEndian.AppendUint32(out, version)
}

// The identity vector wraps the key 201, 202, 203… for itself, version 1, with the ephemeral secret 1, 1, 1… and the
// nonce 9, 9, 9….
func TestTheDocumentedWrapIsMadeTheSameHere(t *testing.T) {
	recipient, err := curve25519.X25519(counting(1, KeyLen), curve25519.Basepoint)
	require.NoError(t, err)
	assert.Equal(t, b64(t, "B6N8vBQgk8i3VdwbEOhstCY3StFqqFPtC9/AsrhtHHw="), recipient)

	ephemeralSecret := repeated(1, KeyLen)
	ephemeral, err := curve25519.X25519(ephemeralSecret, curve25519.Basepoint)
	require.NoError(t, err)
	assert.Equal(t, b64(t, "pOCSkrZRwni5dyxWn1+puxPZBrRqtoyd+dwrRAn4ogk="), ephemeral)
	shared, err := curve25519.X25519(ephemeralSecret, recipient)
	require.NoError(t, err)
	assert.Equal(t, b64(t, "n7N3S7Rf6yLXUjr7Rik1/xE7woA8DndtxaUMe1DC7DQ="), shared)

	info := append([]byte("shellwing collection key v1"), bound(vectorCollection, 1)...)
	sealing := make([]byte, KeyLen)
	_, err = io.ReadFull(hkdf.New(sha256.New, shared, append(append([]byte(nil), ephemeral...), recipient...), info),
		sealing)
	require.NoError(t, err)
	assert.Equal(t, b64(t, "Rz1rtIlGoxq+04hvepUDwsOw4JxQCgXPn8thLKcpb0Y="), sealing)

	aead, err := chacha20poly1305.NewX(sealing)
	require.NoError(t, err)
	key, nonce := counting(201, KeyLen), repeated(9, chacha20poly1305.NonceSizeX)
	sealed := aead.Seal(append([]byte(nil), nonce...), nonce, key, info)
	assert.Equal(t, b64(t, "CQkJCQkJCQkJCQkJCQkJCQkJCQkJCQkJ"+
		"D3nIsVTflOvUIw0qRWj6CBdhXmf5+zLW2+BQ21NL3D9QFo166bTPj7YqYRaS5Ymk"), sealed)
	commitment := sha256.Sum256(append([]byte("shellwing collection key commitment v1"), key...))
	assert.Equal(t, b64(t, "ruUTPgQFKROEd2yEBlBz55dKRs3Hj/1Lnb8OrpqrBE0="), commitment[:])

	w := &Wrap{
		CollectionID:    vectorCollection,
		Version:         1,
		RecipientX25519: recipient,
		EphemeralPublic: ephemeral,
		Sealed:          sealed,
		Commitment:      commitment[:],
	}
	wrapper := ed25519.NewKeyFromSeed(counting(33, ed25519.SeedSize))
	signature := ed25519.Sign(wrapper, w.Signed())
	assert.Equal(t, b64(t, "2VxS/LEyomoRmfqlvvyPEQi0jSFkkUZU63a8E+FKTL6OECH8jAx9UHezgF668S0ISYBwxRcFXOYLNgxDtHqWCw=="),
		signature)
	assert.True(t, w.Verify(wrapper.Public().(ed25519.PublicKey), signature), "what the server takes")
}

// A collection's record, sealed with the key 201, 202, 203…, version 1, and the nonce 7, 7, 7…: a cipher id, the
// nonce, then the ciphertext, with the record's id, clear kind, collection and version as additional data.
func TestTheDocumentedRecordIsSealedTheSameHere(t *testing.T) {
	plain := []byte(`{"v":1,"type":"host","name":"db","f":{"address":{"val":"10.0.0.9","t":"1792312831000-0000-dev1"}}}`)
	aad := []byte("c0110000000000000000000000000001\x00host\x00" + vectorCollection + "\x001")
	aead, err := chacha20poly1305.NewX(counting(201, KeyLen))
	require.NoError(t, err)
	nonce := repeated(7, chacha20poly1305.NonceSizeX)
	sealed := append([]byte{1}, nonce...)
	sealed = aead.Seal(sealed, nonce, plain, aad)
	assert.Equal(t, b64(t, "AQcHBwcHBwcHBwcHBwcHBwcHBwcHBwcHBz3cVlKzDzIoGY6n90kcS7luilmfa80P8cBXS+cwiUEtSKAj"+
		"E8UnX/sSTJphbM9dG4haPwAW3e9pGIPT2DJRRMXuW++aWOHj5AG4yIRBuaTGQO1cjEvbqk02T+9mtifTGSa+m8lsrCoNT3oJ+K8qo6Nrbw=="),
		sealed)
}
