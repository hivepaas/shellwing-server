// Package sshca is an SSH certificate authority: it makes CA keys and signs short-lived user certificates. It holds no
// state, and what a certificate may do is its callers' to decide.
package sshca

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/hivepaas/shellwing-server/hperrors"
)

const (
	// clockSkew backdates a certificate, for a target server whose clock is a little behind this one's.
	clockSkew = 5 * time.Minute
	// MaxPublicKeyLen bounds the text taken as a public key: the largest kind in use is a few kilobytes.
	MaxPublicKeyLen = 16 << 10
	// MaxPrincipals bounds how many accounts one person is granted.
	MaxPrincipals = 32

	minRSABits = 2048
)

// principalPattern admits what account names are. It leaves out what would change meaning where principals are written
// down: commas split them in a certificate's option and in sshd's configuration, and spaces and # in its files.
var principalPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.@-]{0,63}$`)

// ValidPrincipal says whether a name can be granted.
func ValidPrincipal(name string) bool {
	return principalPattern.MatchString(name)
}

// The extensions a certificate may carry. Each one is a thing sshd otherwise refuses: an extension absent from a
// certificate is refused whatever sshd's own configuration allows, so this is the narrower of the two and the one
// worth getting right.
const (
	// ExtensionPermitPTY allows a terminal, which an interactive shell needs and nothing else does.
	ExtensionPermitPTY = "permit-pty"
	// ExtensionPermitPortForwarding allows -L, -R and -D. It is how a person reaches a machine they were never
	// granted, through one they were, so it is off unless an organization turns it on.
	ExtensionPermitPortForwarding = "permit-port-forwarding"
	// ExtensionPermitAgentForwarding allows the target server to use the keys in the person's agent. Anybody with
	// root there can use them too, for as long as the session lasts.
	ExtensionPermitAgentForwarding = "permit-agent-forwarding"
	// ExtensionPermitX11Forwarding allows X11, which carries input and the screen both ways.
	ExtensionPermitX11Forwarding = "permit-X11-forwarding"
	// ExtensionPermitUserRC runs the target account's ~/.ssh/rc on login.
	ExtensionPermitUserRC = "permit-user-rc"
)

// AllExtensions is every extension that may be asked for, in the order a reader expects them.
var AllExtensions = []string{
	ExtensionPermitPTY,
	ExtensionPermitPortForwarding,
	ExtensionPermitAgentForwarding,
	ExtensionPermitX11Forwarding,
	ExtensionPermitUserRC,
}

// DefaultExtensions is what an organization gets until it says otherwise: a terminal, and nothing that reaches past
// the machine being logged in to. `ssh-keygen -s` grants all five by default, which is the convenience of a person
// signing one certificate for themselves rather than a server signing them for everybody.
var DefaultExtensions = []string{ExtensionPermitPTY}

// ValidExtension says whether the name is an extension that may be asked for.
func ValidExtension(name string) bool {
	return slices.Contains(AllExtensions, name)
}

// CAKey is a new CA's key pair: the public half as a line for TrustedUserCAKeys, the private half as PEM.
type CAKey struct {
	PublicKey  string
	PrivateKey string
}

// GenerateCA makes an Ed25519 CA key.
func GenerateCA() (*CAKey, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	sshPublic, err := ssh.NewPublicKey(public)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	block, err := ssh.MarshalPrivateKey(private, "")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &CAKey{
		PublicKey:  strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPublic))),
		PrivateKey: string(pem.EncodeToMemory(block)),
	}, nil
}

// ParseUserKey reads a public key as authorized_keys has it, and refuses a certificate, or a kind or size not signed.
func ParseUserKey(text string) (ssh.PublicKey, error) {
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(strings.TrimSpace(text))) //nolint:dogsled // only the key
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSSHPublicKeyInvalid)
	}
	switch key.Type() {
	case ssh.KeyAlgoED25519, ssh.KeyAlgoECDSA256, ssh.KeyAlgoECDSA384, ssh.KeyAlgoECDSA521,
		ssh.KeyAlgoSKED25519, ssh.KeyAlgoSKECDSA256:
		return key, nil
	case ssh.KeyAlgoRSA:
		crypto, isCrypto := key.(ssh.CryptoPublicKey)
		if !isCrypto {
			return nil, hperrors.Wrap(hperrors.ErrSSHPublicKeyInvalid)
		}
		public, isRSA := crypto.CryptoPublicKey().(*rsa.PublicKey)
		if !isRSA || public.N.BitLen() < minRSABits {
			return nil, hperrors.Wrap(hperrors.ErrSSHPublicKeyInvalid)
		}
		return key, nil
	}
	return nil, hperrors.Wrap(hperrors.ErrSSHPublicKeyInvalid)
}

// UserCert is what to sign.
type UserCert struct {
	Key        ssh.PublicKey
	KeyID      string
	Principals []string
	// Extensions is what the certificate is allowed to do, from AllExtensions. None is allowed and means a
	// certificate that logs in and runs a command, with no terminal.
	Extensions []string
	TTL        time.Duration
	Now        time.Time
}

// Signed is a certificate, as a line for a `-cert.pub` file, and what it says.
type Signed struct {
	Certificate string
	Serial      uint64
	Fingerprint string
	ValidAfter  time.Time
	ValidBefore time.Time
}

// SignerFromPEM reads a CA private key this process holds, as GenerateCA returns one.
//
// It is the other way of getting a Signer, beside a handle to a key a service holds and never hands over. Both end up
// as an ssh.Signer, which is why Sign takes one rather than a key: it is the narrowest thing signing needs, and it is
// the seam that decides whether this process can be made to sign with a key after it has been taken from it.
func SignerFromPEM(caPrivateKey string) (ssh.Signer, error) {
	signer, err := ssh.ParsePrivateKey([]byte(caPrivateKey))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return signer, nil
}

// PublicKeyLine is the signer's public half, as a line for TrustedUserCAKeys.
//
// For a key held elsewhere this is the only way to write it down: there is no private key here to derive it from, and
// what the service answers with is the public half alone.
func PublicKeyLine(signer ssh.Signer) string {
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey())))
}

// Ed25519PublicKeyOf reads the raw public key out of a line PublicKeyLine wrote.
//
// It is how a signer for a key held elsewhere is built without asking the service
// anything: the public half is already written down beside the authority, which
// is what the target servers were told to trust, so signing costs one call to the
// service rather than two.
func Ed25519PublicKeyOf(line string) (ed25519.PublicKey, error) {
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(strings.TrimSpace(line))) //nolint:dogsled // only the key
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSSHPublicKeyInvalid)
	}
	crypto, isCrypto := key.(ssh.CryptoPublicKey)
	if !isCrypto {
		return nil, hperrors.Wrap(hperrors.ErrSSHPublicKeyInvalid)
	}
	public, isEd25519 := crypto.CryptoPublicKey().(ed25519.PublicKey)
	if !isEd25519 {
		return nil, hperrors.Wrap(hperrors.ErrSSHPublicKeyInvalid)
	}
	return public, nil
}

// Sign signs a user certificate. A certificate with no principals would be good for every account, so one is refused.
// What it is allowed to do is its caller's to decide, and only what AllExtensions names can be asked for: an unknown
// extension would be carried into the certificate and mean nothing to sshd, which is worse than being refused here.
//
// The signer may be a key this process holds or one a service holds; signing is the same either way, and nothing below
// this line knows which it got.
func Sign(signer ssh.Signer, cert *UserCert) (*Signed, error) {
	if len(cert.Principals) == 0 {
		return nil, hperrors.Wrap(hperrors.ErrSSHNoPrincipals)
	}
	// MaxPrincipals is checked here rather than only where a grant is made: a member's own accounts and each of
	// their groups' are bounded separately, and it is their union that ends up in a certificate.
	if len(cert.Principals) > MaxPrincipals {
		return nil, hperrors.Wrap(hperrors.ErrSSHTooManyPrincipals)
	}
	extensions := map[string]string{}
	for _, name := range cert.Extensions {
		if !ValidExtension(name) {
			return nil, hperrors.Wrap(hperrors.ErrSSHExtensionInvalid).WithParam("Name", name)
		}
		// An extension is a name with an empty value: what it allows is in the name.
		extensions[name] = ""
	}
	serial, err := newSerial()
	if err != nil {
		return nil, err
	}
	validAfter := cert.Now.Add(-clockSkew).Truncate(time.Second)
	validBefore := cert.Now.Add(cert.TTL).Truncate(time.Second)
	sshCert := &ssh.Certificate{
		Key:             cert.Key,
		Serial:          serial,
		CertType:        ssh.UserCert,
		KeyId:           cert.KeyID,
		ValidPrincipals: cert.Principals,
		ValidAfter:      uint64(validAfter.Unix()),  //nolint:gosec // G115: a time after 1970
		ValidBefore:     uint64(validBefore.Unix()), //nolint:gosec // G115: a time after 1970
		Permissions:     ssh.Permissions{Extensions: extensions},
	}
	if err = sshCert.SignCert(rand.Reader, signer); err != nil {
		return nil, hperrors.Wrap(fmt.Errorf("signing a certificate: %w", err))
	}
	return &Signed{
		Certificate: strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshCert))),
		Serial:      serial,
		Fingerprint: ssh.FingerprintSHA256(cert.Key),
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
	}, nil
}

// newSerial is random, within 63 bits: sshd's logs and JSON readers are not all good for the full 64.
func newSerial() (uint64, error) {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return 0, hperrors.Wrap(err)
	}
	return binary.BigEndian.Uint64(data[:]) >> 1, nil
}
