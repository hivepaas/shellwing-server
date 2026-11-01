package useruc

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/config"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/database"
	"github.com/hivepaas/shellwing-server/pkg/auditdetail"
	"github.com/hivepaas/shellwing-server/pkg/bunex"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
	"github.com/hivepaas/shellwing-server/pkg/transaction"
	"github.com/hivepaas/shellwing-server/pkg/ulid"
	"github.com/hivepaas/shellwing-server/service/emailservice"
	"github.com/hivepaas/shellwing-server/service/userservice"
	"github.com/hivepaas/shellwing-server/usecase/useruc/userdto"
)

func (uc *UC) InviteUser(
	ctx context.Context,
	auth *basedto.Auth,
	req *userdto.InviteUserReq,
) (*userdto.InviteUserResp, error) {
	inviteData := &userInviteData{}
	err := uc.loadUserInviteData(ctx, uc.db, req, inviteData)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	persistingData := &userservice.PersistingUserData{}
	uc.preparePersistingUserInviteData(req, inviteData, persistingData)

	err = transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		err := uc.authorizeAccessChanges(ctx, db, auth, inviteData.User,
			accessResourceTypesToReplace(true, true), persistingData)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if err = uc.userService.PersistUserData(ctx, db, persistingData); err != nil {
			return hperrors.Wrap(err)
		}

		// The role goes in with its value. It is what the account may do from the
		// moment it exists, it is not a secret, and "who let an admin in" is the
		// question this entry is kept for.
		return uc.recordUserChange(ctx, db, auth, base.AuditLogTypeUserCreate,
			auditSectionInvite, inviteData.User, auditdetail.New().
				Set("role", inviteData.User.Role).
				Set("securityOption", inviteData.User.SecurityOption).
				Set("emailSent", req.SendInviteEmail))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if req.SendInviteEmail {
		err = uc.emailService.SendMailUserInvite(ctx, uc.db, &emailservice.EmailDataUserInvite{
			BaseTemplateData: emailservice.BaseTemplateData{
				Email:      inviteData.SystemEmail,
				Recipients: []string{inviteData.User.Email},
				Subject:    "You’ve been invited to join shellwing",
			},
			InviterName:    gofn.Coalesce(auth.User.FullName, auth.User.Username),
			UserSignupLink: inviteData.InviteLink,
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	return &userdto.InviteUserResp{
		Data: &userdto.InviteUserDataResp{
			InviteLink: inviteData.InviteLink,
		},
	}, nil
}

type userInviteData struct {
	User        *entity.User
	InviteLink  string
	SystemEmail *entity.Email
}

func (uc *UC) loadUserInviteData(
	ctx context.Context,
	db database.IDB,
	req *userdto.InviteUserReq,
	data *userInviteData,
) error {
	// A pending user being re-invited has grants this invite replaces.
	user, err := uc.userRepo.GetByEmail(ctx, db, req.Email,
		bunex.SelectRelation("Accesses"),
	)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	if user != nil && user.Status != base.UserStatusPending {
		return hperrors.NewAlreadyExist("User").
			WithMsgLog("user '%s' already exists", req.Email)
	}

	if req.SendInviteEmail {
		emailSetting, err := uc.emailService.GetDefaultSystemEmail(ctx, uc.db)
		if err != nil {
			return hperrors.Wrap(err)
		}
		email, err := emailSetting.AsEmail()
		if err != nil {
			return hperrors.Wrap(err)
		}
		data.SystemEmail = email
	}

	username, err := userservice.UsernameForEmail(ctx, db, uc.userRepo, req.Email)
	if err != nil {
		return hperrors.Wrap(err)
	}

	if user == nil {
		user = &entity.User{
			ID:        gofn.Must(ulid.NewStringULID()),
			Username:  username,
			Email:     req.Email,
			FullName:  username,
			CreatedAt: time.Now(),
		}
	}
	data.User = user

	// Generate invite token
	inviteToken, err := uc.userService.GenerateUserInviteToken(user.ID)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.InviteLink = config.Current().DashboardUserSignupURL(inviteToken)

	return nil
}

func (uc *UC) preparePersistingUserInviteData(
	req *userdto.InviteUserReq,
	data *userInviteData,
	persistingData *userservice.PersistingUserData,
) {
	timeNow := timeutil.NowUTC()
	user := data.User

	user.Role = req.Role
	user.Status = base.UserStatusPending
	user.SecurityOption = req.SecurityOption
	user.AccessExpireAt = req.AccessExpireAt
	user.UpdatedAt = timeNow

	persistingData.UpsertingUsers = append(persistingData.UpsertingUsers, user)

	uc.preparePersistingUserModuleAccesses(user, req.ModuleAccesses, timeNow, persistingData)
	uc.preparePersistingUserCapabilities(user, req.Capabilities, timeNow, persistingData)
}

func (uc *UC) preparePersistingUserModuleAccesses(
	user *entity.User,
	moduleReqs basedto.ModuleAccessSliceReq,
	timeNow time.Time,
	persistingData *userservice.PersistingUserData,
) {
	for _, moduleReq := range moduleReqs {
		persistingData.UpsertingAccesses = append(persistingData.UpsertingAccesses,
			&entity.ACLPermission{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    user.ID,
				ResourceType: base.ResourceTypeModule,
				ResourceID:   moduleReq.ID,
				Actions:      moduleReq.Access,
				CreatedAt:    timeNow,
				UpdatedAt:    timeNow,
			})
	}
}

// preparePersistingUserCapabilities turns the requested capabilities into ACL
// rows. A capability is a single operation, so the row carries exactly the action
// the capability check asks for - see permissionimpl.capabilityAction.
func (uc *UC) preparePersistingUserCapabilities(
	user *entity.User,
	capabilityReqs basedto.CapabilitySliceReq,
	timeNow time.Time,
	persistingData *userservice.PersistingUserData,
) {
	for _, capReq := range capabilityReqs {
		persistingData.UpsertingAccesses = append(persistingData.UpsertingAccesses,
			&entity.ACLPermission{
				SubjectType:  base.SubjectTypeUser,
				SubjectID:    user.ID,
				ResourceType: base.ResourceTypeCapability,
				ResourceID:   string(capReq),
				Actions:      base.AccessActions{Exec: true},
				CreatedAt:    timeNow,
				UpdatedAt:    timeNow,
			})
	}
}
