package collectionuc

import (
	"bytes"
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/collectionkey"
	"github.com/hivepaas/shellwing-server/pkg/identity"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/usecase/collectionuc/collectiondto"
)

// GetKeys returns the collection's current key version and the person's wraps of its keys, for a person it is served
// to. Each fetch is in the audit log: it is what opens the collection.
func (uc *UC) GetKeys(ctx context.Context, auth *basedto.Auth, orgID, collectionID string) (
	*collectiondto.GetKeysResp, error) {
	c, err := uc.reach(ctx, uc.db, auth, orgID, collectionID, false)
	if err != nil {
		return nil, err
	}
	if c.permission == "" {
		return nil, hperrors.NewForbidden("Collection")
	}
	wraps, err := uc.collectionRepo.ListWraps(ctx, uc.db, collectionID, auth.UserID())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data := &collectiondto.KeysResp{KeyVersion: c.collection.KeyVersion, Wraps: []*collectiondto.WrapResp{}}
	versions := make([]int, 0, len(wraps))
	for _, w := range wraps {
		data.Wraps = append(data.Wraps, &collectiondto.WrapResp{
			Version:            w.KeyVersion,
			RecipientX25519:    w.RecipientX25519,
			EphemeralPublic:    w.EphemeralPublic,
			SealedKey:          w.SealedKey,
			Commitment:         w.Commitment,
			Signature:          w.Signature,
			WrappedBy:          w.WrappedBy,
			WrapperUsername:    w.WrapperUsername,
			WrapperX25519:      w.WrapperX25519,
			WrapperEd25519:     w.WrapperEd25519,
			WrapperFingerprint: identity.Fingerprint(w.WrapperX25519, w.WrapperEd25519),
		})
		versions = append(versions, w.KeyVersion)
	}
	if len(wraps) > 0 {
		err = uc.record(ctx, uc.db, auth, c.collection, base.AuditLogTypeCollectionKeyFetch, "",
			auditdetail.New().Set("versions", versions))
		if err != nil {
			return nil, err
		}
	}
	return &collectiondto.GetKeysResp{Data: data}, nil
}

// CreateKey takes a new key for the collection, version current + 1, as an app with Manage made it: the first, or one
// replacing the current, which that app must hold. It is wrapped for the caller and every owner with an identity, and
// may be for anybody else with access; the others wait for it.
func (uc *UC) CreateKey(ctx context.Context, auth *basedto.Auth, orgID, collectionID string,
	req *collectiondto.WrapsReq) (*collectiondto.WrapsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		c, s, mine, err := uc.prepareWraps(ctx, tx, auth, orgID, collectionID)
		if err != nil {
			return err
		}
		if c.permission != base.CollectionPermissionManage {
			return hperrors.NewForbidden("Collection")
		}
		current := c.collection.KeyVersion
		if req.Version != current+1 {
			return hperrors.Wrap(hperrors.ErrCollectionKeyVersion)
		}
		if current > 0 && !s.holders[auth.UserID()] {
			return hperrors.Wrap(hperrors.ErrCollectionNoKey)
		}
		wraps, err := uc.checkWraps(ctx, tx, c.collection, req, mine, s, false)
		if err != nil {
			return err
		}
		if err = uc.requireWrapsFor(ctx, tx, auth, wraps, s); err != nil {
			return err
		}
		if err = uc.collectionRepo.InsertWraps(ctx, tx, wraps); err != nil {
			return hperrors.Wrap(err)
		}
		if err = uc.collectionRepo.SetKeyVersion(ctx, tx, orgID, collectionID, req.Version); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.record(ctx, tx, auth, c.collection, base.AuditLogTypeCollectionUpdate, auditSectionKey,
			auditdetail.New().Set("version", req.Version).Set("recipients", recipients(wraps, s)))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &collectiondto.WrapsResp{}, nil
}

// AddWraps gives the current key to people with access who wait for it, as the app of somebody who manages the
// collection's access and holds the key wrapped it, once they compared the fingerprints.
func (uc *UC) AddWraps(ctx context.Context, auth *basedto.Auth, orgID, collectionID string,
	req *collectiondto.WrapsReq) (*collectiondto.WrapsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		c, s, mine, err := uc.prepareWraps(ctx, tx, auth, orgID, collectionID)
		if err != nil {
			return err
		}
		if !c.managesAccess() {
			return hperrors.NewForbidden("Collection")
		}
		if c.collection.KeyVersion == 0 || req.Version != c.collection.KeyVersion {
			return hperrors.Wrap(hperrors.ErrCollectionKeyVersion)
		}
		if !s.holders[auth.UserID()] {
			return hperrors.Wrap(hperrors.ErrCollectionNoKey)
		}
		wraps, err := uc.checkWraps(ctx, tx, c.collection, req, mine, s, true)
		if err != nil {
			return err
		}
		if err = uc.collectionRepo.InsertWraps(ctx, tx, wraps); err != nil {
			return hperrors.Wrap(err)
		}
		return uc.record(ctx, tx, auth, c.collection, base.AuditLogTypeCollectionUpdate, auditSectionWrap,
			auditdetail.New().Set("version", req.Version).Set("recipients", recipients(wraps, s)))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &collectiondto.WrapsResp{}, nil
}

// prepareWraps locks the collection for a person it is served to, and gathers who has access, who holds the key, and
// the person's identity, which signs their wraps.
func (uc *UC) prepareWraps(ctx context.Context, tx database.IDB, auth *basedto.Auth, orgID, collectionID string) (
	*caller, *standing, *entity.UserIdentity, error) {
	c, err := uc.reach(ctx, tx, auth, orgID, collectionID, true)
	if err != nil {
		return nil, nil, nil, err
	}
	if c.permission == "" {
		return nil, nil, nil, hperrors.NewForbidden("Collection")
	}
	standings, err := uc.standings(ctx, tx, orgID, collectionID)
	if err != nil {
		return nil, nil, nil, err
	}
	s := standings[collectionID]
	if s == nil {
		s = &standing{entitled: map[string]*entity.CollectionEntitlement{}, holders: map[string]bool{}}
	}
	mine, err := uc.identityRepo.Get(ctx, tx, auth.UserID())
	if err != nil {
		return nil, nil, nil, hperrors.Wrap(err)
	}
	if mine == nil {
		return nil, nil, nil, hperrors.Wrap(hperrors.ErrIdentityMissing)
	}
	return c, s, mine, nil
}

// checkWraps checks each wrap is for somebody with access, once, for their identity's key, and signed by the caller's
// identity. fresh: they must not hold this version yet.
func (uc *UC) checkWraps(ctx context.Context, db database.IDB, collection *entity.Collection,
	req *collectiondto.WrapsReq, mine *entity.UserIdentity, s *standing, fresh bool) (
	[]*entity.CollectionKeyWrap, error) {
	now := timeutil.NowUTC()
	seen := map[string]bool{}
	wraps := make([]*entity.CollectionKeyWrap, 0, len(req.Wraps))
	for _, w := range req.Wraps {
		invalid := hperrors.Wrap(hperrors.ErrCollectionWrapInvalid).WithParam("Name", w.UserID)
		entitled := s.entitled[w.UserID]
		if entitled == nil || seen[w.UserID] || (fresh && s.holders[w.UserID]) {
			return nil, invalid
		}
		seen[w.UserID] = true
		invalid = invalid.WithParam("Name", entitled.Username)
		theirs, err := uc.identityRepo.Get(ctx, db, w.UserID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if theirs == nil || !bytes.Equal(theirs.X25519Public, w.RecipientX25519) {
			return nil, invalid
		}
		wrap := &collectionkey.Wrap{
			CollectionID:    collection.ID,
			Version:         req.Version,
			RecipientX25519: w.RecipientX25519,
			EphemeralPublic: w.EphemeralPublic,
			Sealed:          w.SealedKey,
			Commitment:      w.Commitment,
		}
		if !wrap.Verify(mine.Ed25519Public, w.Signature) {
			return nil, invalid
		}
		wraps = append(wraps, &entity.CollectionKeyWrap{
			CollectionID:    collection.ID,
			UserID:          w.UserID,
			KeyVersion:      req.Version,
			RecipientX25519: w.RecipientX25519,
			EphemeralPublic: w.EphemeralPublic,
			SealedKey:       w.SealedKey,
			Commitment:      w.Commitment,
			WrappedBy:       mine.UserID,
			WrapperX25519:   mine.X25519Public,
			WrapperEd25519:  mine.Ed25519Public,
			Signature:       w.Signature,
			CreatedAt:       now,
		})
	}
	return wraps, nil
}

// requireWrapsFor checks a new key is wrapped for the caller, and for every owner with an identity: otherwise a
// collection whose members all left would be lost.
func (uc *UC) requireWrapsFor(ctx context.Context, db database.IDB, auth *basedto.Auth,
	wraps []*entity.CollectionKeyWrap, s *standing) error {
	wrapped := map[string]bool{}
	for _, w := range wraps {
		wrapped[w.UserID] = true
	}
	if !wrapped[auth.UserID()] {
		return hperrors.Wrap(hperrors.ErrCollectionWrapsMissing).WithParam("Name", s.nameOf(auth.UserID()))
	}
	for userID, e := range s.entitled {
		if e.Role != base.OrgRoleOwner || wrapped[userID] {
			continue
		}
		theirs, err := uc.identityRepo.Get(ctx, db, userID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if theirs != nil {
			return hperrors.Wrap(hperrors.ErrCollectionWrapsMissing).WithParam("Name", e.Username)
		}
	}
	return nil
}

func (s *standing) nameOf(userID string) string {
	if e := s.entitled[userID]; e != nil {
		return e.Username
	}
	return userID
}

func recipients(wraps []*entity.CollectionKeyWrap, s *standing) []string {
	out := make([]string, 0, len(wraps))
	for _, w := range wraps {
		out = append(out, s.nameOf(w.UserID))
	}
	return out
}
