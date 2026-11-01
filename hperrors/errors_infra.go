package hperrors

// Errors for files
var (
	ErrFileScopeUnsupported     = NewErr(ErrUnsupported, "ERR_FILE_SCOPE_UNSUPPORTED")
	ErrArchiveFormatUnsupported = NewErr(ErrUnsupported, "ERR_ARCHIVE_FORMAT_UNSUPPORTED")
)

// nolint Errors from infrastructure
var (
	ErrInfra                   = NewErr(ErrInternal, "ERR_INFRA")
	ErrInfraUnknown            = NewErr(ErrInternal, "ERR_INFRA_UNKNOWN")
	ErrInfraInternal           = NewErr(ErrInternal, "ERR_INFRA_INTERNAL")
	ErrInfraInvalidArgument    = NewErr(ErrArgumentInvalid, "ERR_INFRA_INVALID_ARGUMENT")
	ErrInfraNotFound           = NewErr(ErrNotFound, "ERR_INFRA_NOT_FOUND")
	ErrInfraAlreadyExists      = NewErr(ErrAlreadyExist, "ERR_INFRA_ALREADY_EXISTS")
	ErrInfraPermissionDenied   = NewErr(ErrUnauthorized, "ERR_INFRA_PERMISSION_DENIED")
	ErrInfraResourceExhausted  = NewErr(ErrPreconditionFailed, "ERR_INFRA_RESOURCE_EXHAUSTED")
	ErrInfraFailedPrecondition = NewErr(ErrPreconditionFailed, "ERR_INFRA_FAILED_PRECONDITION")
	ErrInfraConflict           = NewErr(ErrConflict, "ERR_INFRA_CONFLICT")
	ErrInfraNotModified        = NewErr(ErrPreconditionFailed, "ERR_INFRA_NOT_MODIFIED")
	ErrInfraAborted            = NewErr(ErrPreconditionFailed, "ERR_INFRA_ABORTED")
	ErrInfraOutOfRange         = NewErr(ErrPreconditionFailed, "ERR_INFRA_OUT_OF_RANGE")
	ErrInfraNotImplemented     = NewErr(ErrNotImplemented, "ERR_INFRA_NOT_IMPLEMENTED")
	ErrInfraUnavailable        = NewErr(ErrUnavailable, "ERR_INFRA_UNAVAILABLE")
	ErrInfraDataLoss           = NewErr(ErrPreconditionFailed, "ERR_INFRA_DATA_LOSS")
	ErrInfraUnauthorized       = NewErr(ErrUnauthorized, "ERR_INFRA_UNAUTHORIZED")
)
