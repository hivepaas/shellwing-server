package orguc

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/usecase/orguc/orgdto"
)

// useCAConfig installs the configuration a new authority's key is made from, and puts back what was there.
func useCAConfig(t *testing.T, cfg config.Config) {
	t.Helper()
	previous := config.Current()
	t.Cleanup(func() { config.SetCurrent(previous) })
	config.SetCurrent(&cfg)
}

// liveVaultCAConfig points new authorities at a real transit engine, or skips.
//
// There is no double here on purpose. A double would answer whatever this code asked it, and what is being checked is
// that a real engine's answers make a certificate every sshd accepts - see pkg/kms for the Vault to start.
func liveVaultCAConfig(t *testing.T) config.Config {
	t.Helper()
	address, token := os.Getenv("SHW_TEST_VAULT_ADDR"), os.Getenv("SHW_TEST_VAULT_TOKEN")
	if address == "" || token == "" {
		t.Skip("SHW_TEST_VAULT_ADDR and SHW_TEST_VAULT_TOKEN not set, skipping the live authority test")
	}
	return config.Config{
		KMS:   config.KMS{VaultAddress: address, VaultToken: token, VaultMount: "transit"},
		SSHCA: config.SSHCA{KeyProvider: base.KMSProviderVaultTransit, VaultKeyPrefix: "shellwing-test-ca-"},
	}
}

// activeCA is the row, to see what was stored rather than what was answered.
func (e *env) activeCA(t *testing.T) *entity.SSHCA {
	t.Helper()
	ca, err := repository.NewSSHCARepo().GetActive(context.Background(), e.db, e.org)
	require.NoError(t, err)
	require.NotNil(t, ca)
	return ca
}

// The whole point: a certificate signed by a key this server cannot read, which the library every sshd speaks for
// accepts, from the authority the hosts were told to trust.
func TestAnAuthorityWhoseKeyIsInTheEngineSignsCertificates(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	useCAConfig(t, liveVaultCAConfig(t))
	admin := e.addMember(t, e.org, base.OrgRoleAdmin, base.OrgMemberStatusActive)
	member := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	_, err := e.uc.SetSSHPrincipals(ctx, admin, e.org, member.UserID(),
		&orgdto.SetSSHPrincipalsReq{Principals: []string{"deploy"}})
	require.NoError(t, err)

	created, err := e.uc.CreateSSHCA(ctx, admin, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)
	assert.Equal(t, base.KMSProviderVaultTransit, created.Data.KeyProvider)
	assert.True(t, strings.HasPrefix(created.Data.PublicKey, "ssh-ed25519 "),
		"an ordinary authority line, so TrustedUserCAKeys takes it unchanged")

	// Nothing of the key is here.
	row := e.activeCA(t)
	assert.Empty(t, row.PrivateKey, "an authority in the engine keeps no private key here")
	assert.False(t, row.KeyIsLocal())
	assert.Equal(t, "shellwing-test-ca-"+row.ID, row.KeyName, "one key per authority, named after it")
	assert.Positive(t, row.KeyVersion, "pinned, so rotating in the engine does not move the authority")

	resp, err := e.uc.IssueSSHCertificate(ctx, member, e.org, issueReq(t))
	require.NoError(t, err)
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(resp.Data.Certificate)) //nolint:dogsled // only the key
	require.NoError(t, err)
	cert, ok := key.(*ssh.Certificate)
	require.True(t, ok)
	assert.Equal(t, []string{"deploy"}, cert.ValidPrincipals)
	assert.Equal(t, created.Data.PublicKey,
		strings.TrimSpace(string(ssh.MarshalAuthorizedKey(cert.SignatureKey))),
		"signed by the authority that was published")

	checker := &ssh.CertChecker{IsUserAuthority: func(auth ssh.PublicKey) bool {
		return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(auth))) == created.Data.PublicKey
	}}
	require.NoError(t, checker.CheckCert("deploy", cert),
		"the signature a real engine made verifies against the authority's published half")
}

// The default is unchanged, because it has to be: an installation with no service to point at must still work.
func TestAnAuthorityKeepsItsKeyHereByDefault(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	useCAConfig(t, config.Config{})
	admin := e.addMember(t, e.org, base.OrgRoleAdmin, base.OrgMemberStatusActive)

	created, err := e.uc.CreateSSHCA(ctx, admin, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)
	assert.Equal(t, base.KMSProviderLocal, created.Data.KeyProvider)

	row := e.activeCA(t)
	assert.True(t, row.KeyIsLocal())
	assert.NotEmpty(t, row.PrivateKey, "sealed, but here")
	assert.Empty(t, row.KeyName)
	assert.Zero(t, row.KeyVersion)
	privateKey, err := row.GetPrivateKey()
	require.NoError(t, err)
	assert.Contains(t, privateKey, "OPENSSH PRIVATE KEY")
}

// An engine that cannot be reached must not leave a half-made authority: an organization whose active authority cannot
// sign can issue nothing at all, and nobody would know until the next person tried to log in.
func TestAnEngineThatCannotBeReachedMakesNoAuthority(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	useCAConfig(t, config.Config{
		// A port nothing listens on, so the attempt fails rather than hangs.
		KMS:   config.KMS{VaultAddress: "http://127.0.0.1:1", VaultToken: "t", VaultMount: "transit"},
		SSHCA: config.SSHCA{KeyProvider: base.KMSProviderVaultTransit, VaultKeyPrefix: "shellwing-test-ca-"},
	})
	admin := e.addMember(t, e.org, base.OrgRoleAdmin, base.OrgMemberStatusActive)

	_, err := e.uc.CreateSSHCA(ctx, admin, e.org, &orgdto.CreateSSHCAReq{})
	require.ErrorIs(t, err, hperrors.ErrSSHCAKeyUnreachable)

	ca, err := repository.NewSSHCARepo().GetActive(ctx, e.db, e.org)
	require.NoError(t, err)
	assert.Nil(t, ca, "and the organization has no authority, rather than one that cannot sign")
}
