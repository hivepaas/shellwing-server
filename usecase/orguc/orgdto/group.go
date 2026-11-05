package orgdto

import (
	"strings"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/shellwing-server/basedto"
	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/pkg/sshca"
)

const (
	groupNameMaxLen        = 100
	groupDescriptionMaxLen = 500
	groupMembersMax        = 1000
)

// GroupRefResp names a group a member is in.
type GroupRefResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GroupMemberResp is a person in a group.
type GroupMemberResp struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

// GroupResp is a group with its members and the SSH accounts it grants each of them.
type GroupResp struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	SSHPrincipals []string           `json:"sshPrincipals"`
	Members       []*GroupMemberResp `json:"members"`
}

func TransformGroup(g *entity.OrgGroup, members []*entity.OrgGroupMemberUser) *GroupResp {
	resp := &GroupResp{
		ID:            g.ID,
		Name:          g.Name,
		Description:   g.Description,
		SSHPrincipals: g.SSHPrincipals,
		Members:       []*GroupMemberResp{},
	}
	if resp.SSHPrincipals == nil {
		resp.SSHPrincipals = []string{}
	}
	for _, m := range members {
		if m.GroupID == g.ID {
			resp.Members = append(resp.Members, &GroupMemberResp{
				UserID: m.UserID, Username: m.Username, FullName: m.FullName,
			})
		}
	}
	return resp
}

type ListGroupsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []*GroupResp  `json:"data"`
}

// GroupReq is a group's name, description and the SSH accounts it grants its members.
type GroupReq struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	SSHPrincipals []string `json:"sshPrincipals"`
}

func NewGroupReq() *GroupReq {
	return &GroupReq{}
}

func (req *GroupReq) ModifyRequest() error {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *GroupReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateStr(&req.Name, true, 1, groupNameMaxLen, "name")
	validators = append(validators, basedto.ValidateStr(&req.Description, false, 0, groupDescriptionMaxLen,
		"description")...)
	validators = append(validators, basedto.ValidateSliceEx(req.SSHPrincipals, false, 0, sshca.MaxPrincipals, nil,
		"sshPrincipals")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GroupDataResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *GroupResp    `json:"data"`
}

type DeleteGroupResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// SetGroupMembersReq makes these members the group's, and nobody else.
type SetGroupMembersReq struct {
	UserIDs []string `json:"userIds"`
}

func NewSetGroupMembersReq() *SetGroupMembersReq {
	return &SetGroupMembersReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *SetGroupMembersReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateSliceEx(req.UserIDs, true, 0, groupMembersMax, nil, "userIds")
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}
