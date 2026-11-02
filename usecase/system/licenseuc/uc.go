package licenseuc

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/service/auditservice"
	"github.com/hivepaas/shellwing-server/service/licenseservice"
	"github.com/hivepaas/shellwing-server/usecase/system/licenseuc/licensedto"
)

const (
	auditSectionInstall = "install"
	auditSectionRemove  = "remove"
)

type UC struct {
	db             *database.DB
	licenseService licenseservice.Service
	auditService   auditservice.Service
}

func New(db *database.DB, licenseService licenseservice.Service, auditService auditservice.Service) *UC {
	return &UC{db: db, licenseService: licenseService, auditService: auditService}
}

func (uc *UC) respOf(ctx context.Context, db database.IDB,
	state *licenseservice.State) (*licensedto.LicenseResp, error) {
	used, err := uc.licenseService.CountUserSeats(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return licensedto.TransformLicense(state, used), nil
}

func (uc *UC) GetLicense(ctx context.Context) (*licensedto.GetLicenseResp, error) {
	state, err := uc.licenseService.Current(ctx, uc.db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data, err := uc.respOf(ctx, uc.db, state)
	if err != nil {
		return nil, err
	}
	return &licensedto.GetLicenseResp{Data: data}, nil
}

func (uc *UC) InstallLicense(ctx context.Context, auth *basedto.Auth,
	req *licensedto.InstallLicenseReq) (*licensedto.InstallLicenseResp, error) {
	var state *licenseservice.State
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		var err error
		state, err = uc.licenseService.Install(ctx, tx, req.Key)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return uc.record(ctx, tx, auth, auditSectionInstall, state)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data, err := uc.respOf(ctx, uc.db, state)
	if err != nil {
		return nil, err
	}
	return &licensedto.InstallLicenseResp{Data: data}, nil
}

func (uc *UC) RemoveLicense(ctx context.Context, auth *basedto.Auth) (*licensedto.RemoveLicenseResp, error) {
	var state *licenseservice.State
	err := transaction.Execute(ctx, uc.db, func(tx database.Tx) error {
		var err error
		state, err = uc.licenseService.Remove(ctx, tx)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return uc.record(ctx, tx, auth, auditSectionRemove, state)
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data, err := uc.respOf(ctx, uc.db, state)
	if err != nil {
		return nil, err
	}
	return &licensedto.RemoveLicenseResp{Data: data}, nil
}

// record keeps who changed the license and what is in force now. Never the key: it holds the customer's email.
func (uc *UC) record(ctx context.Context, db database.IDB, auth *basedto.Auth, section string,
	state *licenseservice.State) error {
	err := auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
		Type:    base.AuditLogTypeLicenseUpdate,
		Scope:   base.ObjectScopeGlobal,
		Source:  base.AuditLogSourceAPIUpdate,
		Section: section,
		Auth:    auth,
		ResType: base.ResourceTypeLicense,
		ResID:   state.InstallationID,
		Detail: auditdetail.New().
			Set("status", state.Status).
			Set("plan", state.Plan).
			Set("maxUsers", state.MaxUsers).String(),
	})
	return hperrors.Wrap(err)
}
