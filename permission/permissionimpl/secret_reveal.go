package permissionimpl

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/permission"
	"github.com/hivepaas/shellwing-server/service/auditservice"
)

// AuthorizeSecretReveal decides whether the caller may see a secret in the clear,
// and records the answer either way.
//
// The record is written before the secret is handed over and its failure aborts
// the reveal. A secret released without a trace is the case this whole path
// exists to prevent, so a reveal that cannot be recorded does not happen.
//
// A refusal is recorded too, which is the half that matters: an admin passes the
// capability gate unconditionally, so the attempt is the only thing there is to
// see. Note this is the one place in the audit that records denials - everywhere
// else permission is settled in the handler, before the usecase runs.
func (p *manager) AuthorizeSecretReveal(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	subject *permission.RevealSubject,
) error {
	if subject == nil {
		return hperrors.NewArgumentInvalid("Reveal subject")
	}
	allowed, denyErr := p.canRevealSecrets(ctx, db, auth)
	return p.recordReveal(ctx, db, auth, subject, allowed, denyErr)
}

// recordReveal writes the audit entry of a reveal decided, then answers the
// decision: a record that cannot be written stops the reveal.
func (p *manager) recordReveal(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
	subject *permission.RevealSubject,
	allowed bool,
	denyErr error,
) error {
	result := base.AuditLogResultAllowed
	if !allowed {
		result = base.AuditLogResultDenied
	}
	err := p.auditService.Record(ctx, db, &auditservice.Entry{
		Type:     base.AuditLogTypeSecretReveal,
		Scope:    subject.Scope,
		ObjectID: subject.ObjectID,
		Source:   subject.Source,
		Result:   result,
		Auth:     auth,
		ResType:  subject.ResType,
		ResID:    subject.ResID,
		ResName:  subject.ResName,
		Detail:   subject.Detail,
	})
	if err != nil {
		return hperrors.Wrap(err)
	}

	return denyErr
}

// canRevealSecrets reports whether the caller may see a secret in the clear.
//
// Two gates, and they answer different questions. The config flag is the
// operator's: it lives in a file on the host, and the one endpoint that can write
// it makes the caller re-enter the app secret, so a session that has taken over an
// admin account cannot turn it on. The capability is the account's. Note that an
// admin passes the second gate unconditionally - see CheckAccess - which is
// precisely why the attempt is recorded rather than merely refused.
func (p *manager) canRevealSecrets(
	ctx context.Context,
	db database.IDB,
	auth *basedto.Auth,
) (allowed bool, err error) {
	if !config.Current().Security.ReturnSecretsViaAPI {
		return false, hperrors.Wrap(hperrors.ErrRevealSecretsDisabled)
	}

	hasPerm, err := p.HasCapability(ctx, db, auth, base.ResourceCapSecretReveal)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if !hasPerm {
		return false, hperrors.Wrap(hperrors.ErrUserNotHavePermissionOnRevealSecrets)
	}
	return true, nil
}
