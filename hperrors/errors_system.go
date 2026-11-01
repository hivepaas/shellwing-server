package hperrors

// System errors
var (
	ErrUpdateVerMismatched = NewErr(ErrMismatch, "ERR_UPDATE_VER_MISMATCHED")
)
