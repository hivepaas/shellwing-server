package entity

import (
	"fmt"

	"github.com/hivepaas/shellwing-server/base"
)

type ObjectScope struct {
	ScopeType base.ObjectScopeType
	UserID    string
	User      *User

	LockScopeObject  bool
	NotRequireActive bool
	NoInherited      bool
}

func (s *ObjectScope) IsGlobalScope() bool {
	return s.ScopeType == base.ObjectScopeGlobal
}

func (s *ObjectScope) IsUserScope() bool {
	return s.ScopeType == base.ObjectScopeUser
}

func (s *ObjectScope) ScopeObjectID() string {
	switch s.ScopeType {
	case base.ObjectScopeUser:
		return s.UserID
	case base.ObjectScopeGlobal:
		return ""
	default:
		return ""
	}
}

func (s *ObjectScope) IsValid() bool {
	if s.ScopeType == base.ObjectScopeGlobal {
		return s.ScopeObjectID() == ""
	}
	return s.ScopeObjectID() != ""
}

func (s *ObjectScope) IsObjectLoaded() bool {
	switch s.ScopeType {
	case base.ObjectScopeUser:
		return s.User != nil
	case base.ObjectScopeGlobal:
		return true
	default:
		return false
	}
}

func (s *ObjectScope) GetBaseURLPath() string {
	switch s.ScopeType {
	case base.ObjectScopeUser:
		return fmt.Sprintf("users/%v", s.User.ID)
	case base.ObjectScopeGlobal:
		return ""
	default:
		return ""
	}
}

func NewObjectScopeGlobal() *ObjectScope {
	return &ObjectScope{ScopeType: base.ObjectScopeGlobal}
}

func NewObjectScopeUser(userID string) *ObjectScope {
	return &ObjectScope{
		ScopeType: base.ObjectScopeUser,
		UserID:    userID,
	}
}
