package orguc

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/usecase/orguc/orgdto"
)

// caKeys is the TrustedUserCAKeys file, as a target server would read it.
func caKeys(t *testing.T, e *env, auth *basedto.Auth) []string {
	t.Helper()
	keys, err := e.uc.SSHCAKeys(context.Background(), auth, e.org)
	require.NoError(t, err)
	text := strings.TrimSpace(string(keys))
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

// signedBy is the authority whose signature a certificate carries.
func signedBy(t *testing.T, certificate string) string {
	t.Helper()
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(certificate)) //nolint:dogsled // only the key
	require.NoError(t, err)
	cert, ok := key.(*ssh.Certificate)
	require.True(t, ok)
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(cert.SignatureKey)))
}

// The first half of rotating: the servers are given the new key before anything is signed with it. Signing straight
// away would refuse everybody whose server had not been updated yet.
func TestANewAuthorityIsPublishedBeforeItSigns(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	person := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	grant(t, e, owner, person, "deploy")

	first, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)
	assert.Equal(t, base.SSHCAStatusActive, first.Data.Status, "the first one has to sign, or nothing can")

	second, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)
	assert.Equal(t, base.SSHCAStatusPublished, second.Data.Status)

	// Trusted by the servers, both of them.
	assert.ElementsMatch(t, []string{first.Data.PublicKey, second.Data.PublicKey}, caKeys(t, e, person))

	// And the one that was signing still is.
	issued, err := e.uc.IssueSSHCertificate(ctx, person, e.org, issueReq(t))
	require.NoError(t, err)
	assert.Equal(t, first.Data.PublicKey, signedBy(t, issued.Data.Certificate))

	// Nothing was refused: publishing is not replacing.
	assert.Equal(t, 0, revokedCount(t, krlSays(t, e, owner, issued.Data.Certificate)))
}

// The second half: the new one takes over, and what the old one signed keeps working because it stays published.
func TestPromotingAPublishedAuthorityMakesItSign(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	person := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	grant(t, e, owner, person, "deploy")

	first, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)
	held, err := e.uc.IssueSSHCertificate(ctx, person, e.org, issueReq(t))
	require.NoError(t, err)
	second, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)

	promoted, err := e.uc.SSHCAStatus(ctx, owner, e.org, second.Data.ID,
		&orgdto.SetSSHCAStatusReq{Status: base.SSHCAStatusActive})
	require.NoError(t, err)
	assert.Equal(t, base.SSHCAStatusActive, promoted.Data.Status)

	// New certificates carry the new authority.
	issued, err := e.uc.IssueSSHCertificate(ctx, person, e.org, issueReq(t))
	require.NoError(t, err)
	assert.Equal(t, second.Data.PublicKey, signedBy(t, issued.Data.Certificate))

	// The old one is published rather than retired, so what it signed is still good and still trusted.
	list, err := e.uc.ListSSHCAs(ctx, owner, e.org)
	require.NoError(t, err)
	byID := map[string]base.SSHCAStatus{}
	for _, ca := range list.Data {
		byID[ca.ID] = ca.Status
	}
	assert.Equal(t, base.SSHCAStatusPublished, byID[first.Data.ID])
	assert.ElementsMatch(t, []string{first.Data.PublicKey, second.Data.PublicKey}, caKeys(t, e, person))
	assert.Equal(t, 0, revokedCount(t, krlSays(t, e, owner, held.Data.Certificate)),
		"nobody holding a certificate from the old authority is turned away")
}

// The third step, and the only one that refuses anything.
func TestRetiringAPublishedAuthorityRefusesWhatItSigned(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	person := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	grant(t, e, owner, person, "deploy")

	first, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)
	held, err := e.uc.IssueSSHCertificate(ctx, person, e.org, issueReq(t))
	require.NoError(t, err)
	second, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)
	_, err = e.uc.SSHCAStatus(ctx, owner, e.org, second.Data.ID,
		&orgdto.SetSSHCAStatusReq{Status: base.SSHCAStatusActive})
	require.NoError(t, err)
	fresh, err := e.uc.IssueSSHCertificate(ctx, person, e.org, issueReq(t))
	require.NoError(t, err)

	retired, err := e.uc.SSHCAStatus(ctx, owner, e.org, first.Data.ID,
		&orgdto.SetSSHCAStatusReq{Status: base.SSHCAStatusRetired})
	require.NoError(t, err)
	assert.Equal(t, base.SSHCAStatusRetired, retired.Data.Status)

	assert.Equal(t, []string{second.Data.PublicKey}, caKeys(t, e, person),
		"a retired authority is no longer trusted")
	// A server may still trust the key until it fetches again, so what it signed goes on the list rather than being
	// left to expire.
	out := krlSays(t, e, owner, held.Data.Certificate, fresh.Data.Certificate)
	assert.Equal(t, 1, revokedCount(t, out), "the old one only: %s", out)
	assert.Contains(t, out, "cert0.pub")

	// And its key is gone.
	var private *string
	require.NoError(t, e.db.NewRaw("SELECT private_key FROM ssh_cas WHERE id = ?",
		first.Data.ID).Scan(ctx, &private))
	assert.Nil(t, private)
}

// The one signing is replaced, not stood down: an organization with nothing that signs can issue nothing, and a
// retired authority does not come back.
func TestOnlyAPublishedAuthorityMoves(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	active, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)

	for _, status := range []base.SSHCAStatus{base.SSHCAStatusActive, base.SSHCAStatusRetired} {
		_, err = e.uc.SSHCAStatus(ctx, owner, e.org, active.Data.ID, &orgdto.SetSSHCAStatusReq{Status: status})
		require.ErrorIs(t, err, hperrors.ErrSSHCANotPublished, string(status))
	}

	published, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)
	_, err = e.uc.SSHCAStatus(ctx, owner, e.org, published.Data.ID,
		&orgdto.SetSSHCAStatusReq{Status: base.SSHCAStatusRetired})
	require.NoError(t, err)
	_, err = e.uc.SSHCAStatus(ctx, owner, e.org, published.Data.ID,
		&orgdto.SetSSHCAStatusReq{Status: base.SSHCAStatusActive})
	require.ErrorIs(t, err, hperrors.ErrSSHCANotPublished, "a retired one does not come back")

	// And a status that is not one of the two is not a request.
	_, err = e.uc.SSHCAStatus(ctx, owner, e.org, published.Data.ID,
		&orgdto.SetSSHCAStatusReq{Status: base.SSHCAStatusPublished})
	require.Error(t, err, "published is what an authority becomes, not something to be set")
}

// Each trusted authority is a key that logs in everywhere, so the set stays small enough to reason about.
func TestOnlySoManyAuthoritiesAreTrustedAtOnce(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)

	for i := range base.MaxTrustedSSHCAs {
		_, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
		require.NoError(t, err, "authority %d", i)
	}
	_, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.ErrorIs(t, err, hperrors.ErrSSHCATooManyTrusted)
	assert.Len(t, caKeys(t, e, owner), base.MaxTrustedSSHCAs)

	// Replacing is still allowed at the bound: it retires the one it replaces, so the set does not grow.
	_, err = e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{Replace: true})
	require.NoError(t, err)
	assert.Len(t, caKeys(t, e, owner), base.MaxTrustedSSHCAs)
}

// A target server fetches the keys the way it fetches the revocation list, with the same token: the two files go to
// the same place and are useless apart.
func TestAServerFetchesTheTrustedKeysWithTheOrganizationsToken(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	created, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)

	made, err := e.uc.CreateSSHKRLToken(ctx, owner, e.org)
	require.NoError(t, err)
	token := made.Data.Token

	keys, err := e.uc.SSHCAKeysForToken(ctx, e.org, token)
	require.NoError(t, err)
	assert.Equal(t, created.Data.PublicKey+"\n", string(keys))

	_, err = e.uc.SSHCAKeysForToken(ctx, e.org, "not-the-token")
	require.ErrorIs(t, err, hperrors.ErrSSHKRLTokenInvalid)
	_, err = e.uc.SSHCAKeysForToken(ctx, e.org, "")
	require.ErrorIs(t, err, hperrors.ErrSSHKRLTokenInvalid)
}

func TestOnlyAnAdminMovesAnAuthority(t *testing.T) {
	e, ctx := newEnv(t), context.Background()
	owner := e.addMember(t, e.org, base.OrgRoleOwner, base.OrgMemberStatusActive)
	member := e.addMember(t, e.org, base.OrgRoleMember, base.OrgMemberStatusActive)
	_, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)
	published, err := e.uc.CreateSSHCA(ctx, owner, e.org, &orgdto.CreateSSHCAReq{})
	require.NoError(t, err)

	_, err = e.uc.SSHCAStatus(ctx, member, e.org, published.Data.ID,
		&orgdto.SetSSHCAStatusReq{Status: base.SSHCAStatusActive})
	require.Error(t, err)

	// But any member reads the keys, as any member reads the revocation list: they are a server's configuration.
	assert.Len(t, caKeys(t, e, member), 2)
}
