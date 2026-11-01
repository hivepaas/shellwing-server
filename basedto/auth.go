package basedto

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/pkg/jwtsession"
)

type User struct {
	*entity.User
	AuthClaims *jwtsession.AuthClaims
}

func (u *User) Entity() *entity.User {
	if u != nil {
		return u.User
	}
	return nil
}

type Auth struct {
	User *User

	// AllowedResources resources which the current user can access to
	AllowedResources map[base.ResourceType][]string
}

func (auth *Auth) UserID() string {
	if auth == nil || auth.User == nil {
		return ""
	}
	return auth.User.ID
}

func (auth *Auth) AllowedUsers(inIDs []string) (allowAll bool, allowed []string) {
	return auth.calcResIntersection(auth.AllowedResources[base.ResourceTypeUser], inIDs)
}

func (auth *Auth) AllowedSettings(inIDs []string) (allowAll bool, allowed []string) {
	return auth.calcResIntersection(auth.AllowedResources[base.ResourceTypeSetting], inIDs)
}

// calcResIntersection calculates the final allowed resource IDs.
//
// checkIDs narrows the answer to the IDs a caller is asking about; without any,
// the question is which objects of the type may be seen at all - what a list
// asks - and the answer is every allowed ID.
func (auth *Auth) calcResIntersection(allowedIDs, checkIDs []string) (allowAll bool, allowed []string) {
	if len(checkIDs) == 0 {
		checkIDs = nil
	}
	if allowedIDs == nil {
		return checkIDs == nil, checkIDs
	}
	if len(allowedIDs) == 0 {
		return false, nil // no access on any object of the resource type
	}
	if checkIDs == nil {
		return false, gofn.ToSet(allowedIDs)
	}
	// checkIDs:   a b c
	// allowedIDs: a b d -> result: a b
	return false, gofn.FilterIN(checkIDs, allowedIDs...)
}
