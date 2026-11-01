package permission

import (
	"context"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/infra/database"
)

type Manager interface {
	CheckAccess(ctx context.Context, db database.IDB, auth *basedto.Auth, check AccessCheck) (bool, error)

	// Check access on a specific setting
	CheckAccessOnSetting(ctx context.Context, db database.IDB, auth *basedto.Auth, check AccessCheck,
		setting *entity.Setting) (bool, error)

	// AuthorizeAccessChanges checks the access changes the acting user wants to
	// apply to another subject - a non-admin may only flip the action bits they
	// hold themselves, both when granting and when revoking - and returns the
	// subject's existing rows the operation may replace.
	AuthorizeAccessChanges(ctx context.Context, db database.IDB, auth *basedto.Auth,
		desired []*entity.ACLPermission, current []*entity.ACLPermission) ([]*entity.ACLPermission, error)

	// HasCapability reports whether the caller holds a capability
	HasCapability(ctx context.Context, db database.IDB, auth *basedto.Auth,
		capability base.ResourceCapability) (bool, error)

	// AuthorizeSecretReveal decides whether the caller may see a secret in the clear,
	// and records the answer either way.
	AuthorizeSecretReveal(ctx context.Context, db database.IDB, auth *basedto.Auth, subject *RevealSubject) error

	// NOTE: this func should be called within a transaction
	UpdateACLPermissions(ctx context.Context, db database.IDB, perms []*entity.ACLPermission) error
	DeleteACLPermissions(ctx context.Context, db database.IDB, perms []*base.PermissionResource) error
	DeleteACLPermissionsBySubjects(ctx context.Context, db database.IDB,
		subjectType base.SubjectType, subjectIDs []string) error
	DeleteACLPermissionsByObjects(ctx context.Context, db database.IDB, objectIDs []string) error
	DeleteACLPermissionsOfUsers(ctx context.Context, db database.IDB, userIDs []string) error
}
