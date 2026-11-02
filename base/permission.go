package base

type SubjectType string

const (
	SubjectTypeUser SubjectType = "user"
)

type ResourceType string

const (
	ResourceTypeAuditLog     ResourceType = "audit-log"
	ResourceTypeCapability   ResourceType = "capability"
	ResourceTypeCloudStorage ResourceType = "cloud-storage"
	ResourceTypeEmail        ResourceType = "email"
	ResourceTypeIMService    ResourceType = "im-service"
	ResourceTypeKeyAuth      ResourceType = "key-auth"
	ResourceTypeLicense      ResourceType = "license"
	ResourceTypeModule       ResourceType = "module"
	ResourceTypeNotification ResourceType = "notification"
	ResourceTypeOAuth        ResourceType = "oauth"
	ResourceTypeSetting      ResourceType = "setting"
	ResourceTypeSysError     ResourceType = "sys-error"
	ResourceTypeTask         ResourceType = "task"
	ResourceTypeUser         ResourceType = "user"
)

type ResourceModule string

const (
	ResourceModuleSettings ResourceModule = "mod::settings"
	ResourceModuleUser     ResourceModule = "mod::user"
	ResourceModuleSystem   ResourceModule = "mod::system"
)

var (
	AllResourceModules = []ResourceModule{ResourceModuleSettings, ResourceModuleUser, ResourceModuleSystem}
)

type ResourceCapability string

const (
	// The name of a permission, not a credential - hence the gosec exclusion.
	ResourceCapSecretReveal ResourceCapability = "cap::secret::reveal" //nolint:gosec // G101
)

var (
	AllResourceCapabilities = []ResourceCapability{ResourceCapSecretReveal}
)

type ActionType string

const (
	ActionTypeRead    ActionType = "read"
	ActionTypeExecute ActionType = "execute"
	ActionTypeWrite   ActionType = "write"
	ActionTypeDelete  ActionType = "delete"
)

var (
	AllActionTypes = []ActionType{ActionTypeRead, ActionTypeExecute, ActionTypeWrite, ActionTypeDelete}
)

type AccessActions struct {
	Read  bool `json:"read"`
	Exec  bool `json:"execute"`
	Write bool `json:"write"`
	Del   bool `json:"delete"`
}

func (a *AccessActions) Equal(other AccessActions) bool {
	return a.Read == other.Read && a.Exec == other.Exec && a.Write == other.Write && a.Del == other.Del
}

func (a *AccessActions) IsFullAccess() bool {
	return a.Read && a.Exec && a.Write && a.Del
}

func (a *AccessActions) IsNoAccess() bool {
	return !a.Read && !a.Exec && !a.Write && !a.Del
}

func (a *AccessActions) GetAllowedActions() (allowed []ActionType) {
	if a.Read {
		allowed = append(allowed, ActionTypeRead)
	}
	if a.Exec {
		allowed = append(allowed, ActionTypeExecute)
	}
	if a.Write {
		allowed = append(allowed, ActionTypeWrite)
	}
	if a.Del {
		allowed = append(allowed, ActionTypeDelete)
	}
	return allowed
}

func (a *AccessActions) Allows(action ActionType) bool {
	if a == nil {
		return false
	}
	switch action {
	case ActionTypeRead:
		return a.Read
	case ActionTypeExecute:
		return a.Exec
	case ActionTypeWrite:
		return a.Write
	case ActionTypeDelete:
		return a.Del
	}
	return false
}

func (a *AccessActions) AllowsAll(actions []ActionType) bool {
	if a == nil {
		return false
	}
	for _, action := range actions {
		if !a.Allows(action) {
			return false
		}
	}
	return true
}

func (a *AccessActions) AllowsAny(actions []ActionType) bool {
	if a == nil {
		return false
	}
	if len(actions) == 0 {
		return true
	}
	for _, action := range actions {
		if a.Allows(action) {
			return true
		}
	}
	return false
}

func (a *AccessActions) Reset(r, x, w, d bool) {
	a.Read = r
	a.Exec = x
	a.Write = w
	a.Del = d
}

func NewAccessActions(r, x, w, d bool) AccessActions {
	return AccessActions{
		Read:  r,
		Exec:  x,
		Write: w,
		Del:   d,
	}
}

func NewFullAccessActions() AccessActions {
	return AccessActions{
		Read:  true,
		Exec:  true,
		Write: true,
		Del:   true,
	}
}

type PermissionResource struct {
	SubjectType  SubjectType
	SubjectID    string
	ResourceType ResourceType
	ResourceID   string
}
