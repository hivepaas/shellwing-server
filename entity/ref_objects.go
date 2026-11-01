package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/base"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/timeutil"
)

func NewRefObjects() *RefObjects {
	return &RefObjects{
		RefSettings: make(map[string]*Setting, 5), //nolint:mnd
		RefUsers:    make(map[string]*User),
	}
}

type RefObjects struct {
	RefSettings map[string]*Setting `json:"settings"`
	RefUsers    map[string]*User    `json:"users"`
}

func (r *RefObjects) AddRefObjects(refObjects *RefObjects) {
	if refObjects == nil {
		return
	}

	for _, refSetting := range refObjects.RefSettings {
		r.RefSettings[refSetting.ID] = refSetting
	}
	for _, refUser := range refObjects.RefUsers {
		r.RefUsers[refUser.ID] = refUser
	}
}

func (r *RefObjects) AddObjectScope(scope *ObjectScope) {
	if scope == nil {
		return
	}
	if scope.User != nil {
		r.RefUsers[scope.User.ID] = scope.User
	}
}

//nolint:gocognit
func (r *RefObjects) GetObjectScope(
	scope base.ObjectScopeType,
	objectID string,
	requireActive bool,
) (*ObjectScope, error) {
	switch scope {
	case base.ObjectScopeUser:
		user := r.RefUsers[objectID]
		if user == nil {
			return nil, hperrors.Wrap(hperrors.ErrUserNotFound).WithParam("Name", objectID)
		}
		if requireActive {
			if user.Status != base.UserStatusActive || user.IsAccessExpired() {
				return nil, hperrors.Wrap(hperrors.ErrUserUnavailable).
					WithParam("Name", gofn.Coalesce(user.FullName, user.Username))
			}
		}
		return user.GetObjectScope(), nil

	case base.ObjectScopeGlobal:
		return NewObjectScopeGlobal(), nil
	}
	return nil, nil
}

type RefObjectIDs struct {
	RefSettingIDs []string `json:"settingIds"`
	RefUserIDs    []string `json:"userIds"`
}

func (r *RefObjectIDs) HasData() bool {
	return len(r.RefSettingIDs) > 0 || len(r.RefUserIDs) > 0
}

func (r *RefObjectIDs) AddRefIDs(refIDs *RefObjectIDs) {
	if refIDs == nil {
		return
	}
	r.RefSettingIDs = append(r.RefSettingIDs, refIDs.RefSettingIDs...)
	r.RefUserIDs = append(r.RefUserIDs, refIDs.RefUserIDs...)
}

func (r *RefObjectIDs) AddScopeObjectIDOfSettings(settings ...*Setting) {
	for _, s := range settings {
		if s.ObjectID == "" {
			continue
		}
		switch s.Scope {
		case base.ObjectScopeUser:
			r.RefUserIDs = append(r.RefUserIDs, s.ObjectID)
		case base.ObjectScopeGlobal:
		}
	}
}

func (r *RefObjectIDs) GetRecursiveRefObjectIDs(refObjects *RefObjects) *RefObjectIDs {
	newRefIDs := &RefObjectIDs{}
	for _, setting := range refObjects.RefSettings {
		newRefIDs.AddRefIDs(setting.MustGetRefObjectIDs())
	}
	res := &RefObjectIDs{}
	for _, settingID := range newRefIDs.RefSettingIDs {
		if !gofn.Contain(r.RefSettingIDs, settingID) {
			res.RefSettingIDs = append(res.RefSettingIDs, settingID)
		}
	}
	for _, userID := range newRefIDs.RefUserIDs {
		if !gofn.Contain(r.RefUserIDs, userID) {
			res.RefUserIDs = append(res.RefUserIDs, userID)
		}
	}
	return res
}

func (r *RefObjectIDs) GetResourceLinks(srcType base.ResourceType, srcID string) []*ResLink {
	resLinks := make([]*ResLink, 0, len(r.RefSettingIDs)+len(r.RefUserIDs))
	timeNow := timeutil.NowUTC()
	for _, refSettingID := range r.RefSettingIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeSetting,
			DstID:     refSettingID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	for _, refUserID := range r.RefUserIDs {
		resLinks = append(resLinks, &ResLink{
			SrcType:   srcType,
			SrcID:     srcID,
			DstType:   base.ResourceTypeUser,
			DstID:     refUserID,
			CreatedAt: timeNow,
			UpdatedAt: timeNow,
		})
	}
	return resLinks
}
