package userservice

import (
	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/pkg/secrethelper"
)

var (
	PasswordRequirements = secrethelper.SecretStrengthRequirements{
		MinLen:             secrethelper.DefaultSecretMinLen,
		MaxLen:             secrethelper.DefaultSecretMaxLen,
		RequiredLowercases: secrethelper.DefaultSecretRequiredLowercases,
		RequiredUppercases: secrethelper.DefaultSecretRequiredUppercases,
		RequiredDigits:     secrethelper.DefaultSecretRequiredDigits,
		RequiredSpecials:   secrethelper.DefaultSecretRequiredSpecials,
		MaxSimilarRun:      secrethelper.DefaultSecretMaxSimilarRun,
		MaxSequenceRun:     secrethelper.DefaultSecretMaxSequenceRun,
	}
)

const (
	SkipCheckingCurrentPassword = ""
)

type PersistingUserData struct {
	UpsertingUsers      []*entity.User
	UpsertingSettings   []*entity.Setting
	UpsertingBinObjects []*entity.BinObject
	UpsertingAccesses   []*entity.ACLPermission
	DeletingAccesses    []*base.PermissionResource
}
