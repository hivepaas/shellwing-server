package sshkrl

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/sshca"
)

// signed is a certificate of its own CA, with the serial the CA chose.
type signed struct {
	ca   *sshca.CAKey
	cert *sshca.Signed
}

func sign(t *testing.T, ca *sshca.CAKey, principal string) *signed {
	t.Helper()
	user, err := sshca.GenerateCA() // its public half serves as a user key
	require.NoError(t, err)
	key, err := sshca.ParseUserKey(user.PublicKey)
	require.NoError(t, err)
	signer, err := sshca.SignerFromPEM(ca.PrivateKey)
	require.NoError(t, err)
	cert, err := sshca.Sign(signer, &sshca.UserCert{
		Key: key, KeyID: "user=alice", Principals: []string{principal}, TTL: time.Hour, Now: time.Now(),
	})
	require.NoError(t, err)
	return &signed{ca: ca, cert: cert}
}

func TestWriteIsTheFormatOpenSSHReads(t *testing.T) {
	keygen, err := exec.LookPath("ssh-keygen")
	if err != nil {
		t.Skip("ssh-keygen not installed, skipping the one test that checks the format against OpenSSH itself")
	}
	ca, err := sshca.GenerateCA()
	require.NoError(t, err)
	revoked, kept := sign(t, ca, "deploy"), sign(t, ca, "deploy")

	krl, err := Write(uint64(time.Now().Unix()), time.Now(), []Authority{ //nolint:gosec // G115: after 1970
		{PublicKey: ca.PublicKey, Serials: []uint64{revoked.cert.Serial}},
	})
	require.NoError(t, err)

	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		return path
	}
	krlPath := filepath.Join(dir, "krl")
	require.NoError(t, os.WriteFile(krlPath, krl, 0o600))

	// -Q asks whether the file revokes these certificates. It says so per line and leaves the exit status non-zero
	// when any of them is revoked, so the lines are what to read.
	out, _ := exec.Command(keygen, "-Q", "-f", krlPath, //nolint:gosec // G204: a path this test wrote
		write("revoked-cert.pub", revoked.cert.Certificate),
		write("kept-cert.pub", kept.cert.Certificate)).CombinedOutput()
	lines := string(out)
	assert.Contains(t, lines, "revoked-cert.pub", "ssh-keygen read the list")
	for _, line := range strings.Split(strings.TrimSpace(lines), "\n") {
		switch {
		case strings.Contains(line, "revoked-cert.pub"):
			assert.Contains(t, line, "REVOKED", "the serial that was revoked: %s", line)
		case strings.Contains(line, "kept-cert.pub"):
			assert.NotContains(t, line, "REVOKED", "another certificate of the same CA: %s", line)
		}
	}
}

func TestWriteRevokesOnlyUnderItsOwnAuthority(t *testing.T) {
	keygen, err := exec.LookPath("ssh-keygen")
	if err != nil {
		t.Skip("ssh-keygen not installed")
	}
	mine, err := sshca.GenerateCA()
	require.NoError(t, err)
	theirs, err := sshca.GenerateCA()
	require.NoError(t, err)
	// The same serial, signed by another authority. A serial means nothing except against the CA that issued it.
	ours := sign(t, mine, "deploy")
	krl, err := Write(1, time.Now(), []Authority{{PublicKey: theirs.PublicKey, Serials: []uint64{ours.cert.Serial}}})
	require.NoError(t, err)

	dir := t.TempDir()
	krlPath := filepath.Join(dir, "krl")
	require.NoError(t, os.WriteFile(krlPath, krl, 0o600))
	certPath := filepath.Join(dir, "cert.pub")
	require.NoError(t, os.WriteFile(certPath, []byte(ours.cert.Certificate), 0o600))

	out, _ := exec.Command(keygen, "-Q", "-f", krlPath, certPath).CombinedOutput() //nolint:gosec // G204: as above
	assert.NotContains(t, string(out), "REVOKED", "another authority's list says nothing about ours")
}

func TestWriteCoversEveryAuthorityAnOrganizationHas(t *testing.T) {
	retired, err := sshca.GenerateCA()
	require.NoError(t, err)
	active, err := sshca.GenerateCA()
	require.NoError(t, err)

	both, err := Write(1, time.Now(), []Authority{
		{PublicKey: retired.PublicKey, Serials: []uint64{1}},
		{PublicKey: active.PublicKey, Serials: []uint64{2, 3}},
	})
	require.NoError(t, err)
	one, err := Write(1, time.Now(), []Authority{{PublicKey: active.PublicKey, Serials: []uint64{2, 3}}})
	require.NoError(t, err)
	assert.Greater(t, len(both), len(one), "a retired authority's certificates are revoked too")
}

func TestWriteWithNothingRevoked(t *testing.T) {
	ca, err := sshca.GenerateCA()
	require.NoError(t, err)
	empty, err := Write(1, time.Now(), nil)
	require.NoError(t, err)
	assert.Equal(t, magic, string(empty[:len(magic)]), "a valid list, revoking nothing")

	// An authority with no serials writes no section of its own: there is nothing to say about it.
	none, err := Write(1, time.Now(), []Authority{{PublicKey: ca.PublicKey}})
	require.NoError(t, err)
	assert.Equal(t, empty, none)

	// Nor does a serial that is only there twice, or zero.
	zero, err := Write(1, time.Now(), []Authority{{PublicKey: ca.PublicKey, Serials: []uint64{0, 0}}})
	require.NoError(t, err)
	assert.Equal(t, empty, zero)
}

func TestWriteIsTheSameBytesForTheSameRevocations(t *testing.T) {
	ca, err := sshca.GenerateCA()
	require.NoError(t, err)
	at := time.Now()
	first, err := Write(7, at, []Authority{{PublicKey: ca.PublicKey, Serials: []uint64{3, 1, 2}}})
	require.NoError(t, err)
	// The same serials, in another order and one of them twice.
	second, err := Write(7, at, []Authority{{PublicKey: ca.PublicKey, Serials: []uint64{2, 3, 1, 2}}})
	require.NoError(t, err)
	assert.Equal(t, first, second, "a server can tell a list that changed from one that did not")
}

func TestWriteRefusesSomethingThatIsNoPublicKey(t *testing.T) {
	_, err := Write(1, time.Now(), []Authority{{PublicKey: "not a key", Serials: []uint64{1}}})
	require.ErrorIs(t, err, hperrors.ErrSSHPublicKeyInvalid)
}

// TestWriteMatchesSSHKeygensOwnOutput writes the same revocation twice, once here and once with ssh-keygen, and
// compares them. The two need not agree - ssh-keygen may pack serials into ranges - but while they do, a change in
// either is worth looking at.
func TestWriteMatchesSSHKeygensOwnOutput(t *testing.T) {
	keygen, err := exec.LookPath("ssh-keygen")
	if err != nil {
		t.Skip("ssh-keygen not installed")
	}
	ca, err := sshca.GenerateCA()
	require.NoError(t, err)
	const serial = 12345

	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pub")
	require.NoError(t, os.WriteFile(caPath, []byte(ca.PublicKey+"\n"), 0o600))
	serialsPath := filepath.Join(dir, "serials.txt")
	require.NoError(t, os.WriteFile(serialsPath, []byte("serial: "+strconv.Itoa(serial)+"\n"), 0o600))
	theirs := filepath.Join(dir, "theirs.krl")
	out, err := exec.Command(keygen, "-k", "-f", theirs, "-s", caPath, serialsPath).CombinedOutput() //nolint:gosec
	require.NoError(t, err, string(out))
	expected, err := os.ReadFile(theirs) //nolint:gosec // G304: a path this test wrote
	require.NoError(t, err)

	// ssh-keygen stamps the time it ran; ours takes the same one so that the rest can be compared.
	const generatedAtOffset = 20
	generatedAt := int64(0)
	for _, b := range expected[generatedAtOffset : generatedAtOffset+8] {
		generatedAt = generatedAt<<8 | int64(b)
	}
	ours, err := Write(0, time.Unix(generatedAt, 0), []Authority{{PublicKey: ca.PublicKey, Serials: []uint64{serial}}})
	require.NoError(t, err)
	assert.Equal(t, expected, ours)
}
