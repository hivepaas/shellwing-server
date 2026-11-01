package hperrors

// Errors for settings
var (
	ErrSettingNotFound              = NewErr(ErrNotFound, "ERR_SETTING_NOT_FOUND")
	ErrObjectScopeInvalid           = NewErr(ErrValueInvalid, "ERR_OBJECT_SCOPE_INVALID")
	ErrInheritedSettingNonUpdatable = NewErr(ErrNonEditable, "ERR_INHERITED_SETTING_NON_UPDATABLE")
	ErrSettingTypeUnsupported       = NewErr(ErrUnsupported, "ERR_SETTING_TYPE_UNSUPPORTED")
	ErrIMServiceUnsupported         = NewErr(ErrUnsupported, "ERR_IM_SERVICE_UNSUPPORTED")
	// ErrCloudStorageConnFailed is a bucket a connection test could not reach:
	// the endpoint, the key or the bucket, as Reason says.
	ErrCloudStorageConnFailed = NewErr(ErrPreconditionFailed, "ERR_CLOUD_STORAGE_CONN_FAILED")

	ErrPasswordNotMeetRequirements  = NewErr(ErrArgumentInvalid, "ERR_PASSWORD_NOT_MEET_REQUIREMENTS")
	ErrPasswordHasWeakSequence      = NewErr(ErrArgumentInvalid, "ERR_PASSWORD_HAS_WEAK_SEQUENCE")
	ErrPasswordTooSimilarToPrevious = NewErr(ErrArgumentInvalid, "ERR_PASSWORD_TOO_SIMILAR_TO_PREVIOUS")
	ErrDataVerNewerThanSystemVer    = NewErr(ErrValueInvalid, "ERR_DATA_VER_NEWER_THAN_SYSTEM_VER")
)
