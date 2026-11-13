// Package sshkrl writes an OpenSSH key revocation list: the file sshd's RevokedKeys names, which refuses a
// certificate by serial however long it still had to run.
//
// It is the other half of a short lifetime. A certificate cannot be taken back once signed, so without this the only
// answer to one that should not be used any more is to wait for it to end. A list the target servers read turns that
// wait into a fetch.
//
// Only the CA's public key is needed to write one, since a serial means nothing except against the authority that
// issued it. That is worth knowing: where the private key lives - this process, a KMS, an HSM - makes no difference
// here.
//
// The format is OpenSSH's own, documented in its PROTOCOL.krl. Only the one section this needs is written, a list of
// serials per authority; sshd reads lists, ranges and bitmaps alike, and a list of a few thousand serials is a few
// tens of kilobytes.
package sshkrl

import (
	"encoding/binary"
	"slices"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/hivepaas/shellwing-server/hperrors"
)

const (
	// magic and formatVersion open every KRL.
	magic         = "SSHKRL\n\x00"
	formatVersion = 1
	// sectionCertificates holds what one authority issued; certSectionSerialList, inside it, holds plain serials.
	sectionCertificates  = 1
	certSectionSerialLst = 0x20
)

// Authority is one certificate authority and the serials of its certificates that are to be refused.
//
// PublicKey is the authorized_keys line, as an organization's CA has it. An organization that has replaced its CA has
// more than one: a certificate the old one signed is still good on a server that still trusts it, so the list covers
// every authority a server may hold, not only the one signing now.
type Authority struct {
	PublicKey string
	Serials   []uint64
}

// Write returns a revocation list.
//
// version is the KRL's own, which `ssh-keygen -k -u` compares when merging one list into another; seconds since the
// epoch serve, being both meaningful and always ahead of the list before. A list with nothing revoked is still
// written, and is valid: a server's configuration names the file whether or not anything is in it yet, and a server
// that cannot read the file it was told to read refuses every certificate.
func Write(version uint64, generatedAt time.Time, authorities []Authority) ([]byte, error) {
	out := &writer{}
	out.raw([]byte(magic))
	out.uint32(formatVersion)
	out.uint64(version)
	out.uint64(uint64(generatedAt.Unix())) //nolint:gosec // G115: a time after 1970
	out.uint64(0)                          // flags, none defined
	out.str(nil)                           // reserved
	out.str(nil)                           // comment

	for _, authority := range authorities {
		serials := cleanSerials(authority.Serials)
		if len(serials) == 0 {
			continue
		}
		key, err := parseKey(authority.PublicKey)
		if err != nil {
			return nil, err
		}

		list := &writer{}
		for _, serial := range serials {
			list.uint64(serial)
		}
		section := &writer{}
		section.str(key.Marshal())
		section.str(nil) // reserved
		section.byte(certSectionSerialLst)
		section.str(list.bytes)

		out.byte(sectionCertificates)
		out.str(section.bytes)
	}
	return out.bytes, nil
}

func parseKey(publicKey string) (ssh.PublicKey, error) {
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(strings.TrimSpace(publicKey))) //nolint:dogsled // only the key
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSSHPublicKeyInvalid)
	}
	return key, nil
}

// cleanSerials sorts the serials and keeps each once, so that the same revocations always write the same bytes and a
// server can tell a list that changed from one that did not. Zero is left out: OpenSSH gives it to a certificate with
// no serial at all, and revoking it would mean something nobody asked for.
func cleanSerials(serials []uint64) []uint64 {
	cleaned := slices.Clone(serials)
	slices.Sort(cleaned)
	cleaned = slices.Compact(cleaned)
	return slices.DeleteFunc(cleaned, func(serial uint64) bool { return serial == 0 })
}

// writer builds the SSH wire encoding the format is written in.
type writer struct {
	bytes []byte
}

func (w *writer) raw(value []byte) {
	w.bytes = append(w.bytes, value...)
}

func (w *writer) byte(value byte) {
	w.bytes = append(w.bytes, value)
}

func (w *writer) uint32(value uint32) {
	w.bytes = binary.BigEndian.AppendUint32(w.bytes, value)
}

func (w *writer) uint64(value uint64) {
	w.bytes = binary.BigEndian.AppendUint64(w.bytes, value)
}

// str writes a length-prefixed string, the encoding every variable-length field in the format uses.
func (w *writer) str(value []byte) {
	w.uint32(uint32(len(value))) //nolint:gosec // G115: a section of this file is never 4GiB
	w.raw(value)
}
