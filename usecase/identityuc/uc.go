package identityuc

import (
	"bytes"
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/identity"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/usecase/identityuc/identitydto"
)

type UC struct {
	db             *database.DB
	identityRepo   repository.UserIdentityRepo
	collectionRepo repository.CollectionRepo
	auditService   auditservice.Service
}

func New(db *database.DB, identityRepo repository.UserIdentityRepo, collectionRepo repository.CollectionRepo,
	auditService auditservice.Service) *UC {
	return &UC{db: db, identityRepo: identityRepo, collectionRepo: collectionRepo, auditService: auditService}
}

// GetMine returns the person's identity, or no data when their app has published none.
func (uc *UC) GetMine(ctx context.Context, auth *basedto.Auth) (*identitydto.GetIdentityResp, error) {
	mine, err := uc.identityRepo.Get(ctx, uc.db, auth.UserID())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &identitydto.GetIdentityResp{Data: identitydto.TransformIdentity(mine)}, nil
}

// PutMine keeps the identity the person's app publishes, once its signature is checked. Publishing the same one again
// changes nothing; another takes its place, and is in the audit log with the one before. The collection keys wrapped
// for the one before go: the new one cannot open them, and the person waits for them again.
func (uc *UC) PutMine(ctx context.Context, auth *basedto.Auth, req *identitydto.PutIdentityReq) (
	*identitydto.PutIdentityResp, error) {
	if !identity.Verify(req.X25519Public, req.Ed25519Public, req.Signature) {
		return nil, hperrors.Wrap(hperrors.ErrIdentityInvalid)
	}
	next := &entity.UserIdentity{
		UserID:        auth.UserID(),
		X25519Public:  req.X25519Public,
		Ed25519Public: req.Ed25519Public,
		Signature:     req.Signature,
		Fingerprint:   identity.Fingerprint(req.X25519Public, req.Ed25519Public),
	}
	var saved *entity.UserIdentity
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		before, err := uc.identityRepo.Get(ctx, tx, auth.UserID())
		if err != nil {
			return hperrors.Wrap(err)
		}
		if before != nil && bytes.Equal(before.X25519Public, next.X25519Public) &&
			bytes.Equal(before.Ed25519Public, next.Ed25519Public) {
			saved = before
			return nil
		}
		if err = uc.identityRepo.Put(ctx, tx, next); err != nil {
			return hperrors.Wrap(err)
		}
		if saved, err = uc.identityRepo.Get(ctx, tx, auth.UserID()); err != nil {
			return hperrors.Wrap(err)
		}
		detail := auditdetail.New().Set("fingerprint", next.Fingerprint)
		if before != nil {
			dropped, err := uc.collectionRepo.DeleteUserWraps(ctx, tx, auth.UserID())
			if err != nil {
				return hperrors.Wrap(err)
			}
			detail.Set("before", before.Fingerprint).Set("wrapsDropped", dropped)
		}
		return hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, tx, &auditservice.Entry{
			Scope:   base.ObjectScopeGlobal,
			Source:  base.AuditLogSourceAPIUpdate,
			Auth:    auth,
			Type:    base.AuditLogTypeUserIdentityUpdate,
			ResType: base.ResourceTypeUserIdentity,
			ResID:   auth.UserID(),
			Detail:  detail.String(),
		}))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &identitydto.PutIdentityResp{Data: identitydto.TransformIdentity(saved)}, nil
}
