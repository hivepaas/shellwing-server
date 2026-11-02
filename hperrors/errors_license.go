package hperrors

// Errors for the limit of users and the license key that lifts it.
//
// A rule that refuses is 412, never 403: the web reads 403 as a permission the person lacks, and none of these is about
// who is asking.
var (
	// ErrPlanLimitReached refuses one more user past the limit. Its param Max is the limit.
	ErrPlanLimitReached = NewErr(ErrPreconditionFailed, "ERR_PLAN_LIMIT_REACHED")
	// ErrLicenseKeyInvalid is a text that is no key of this product, or one whose signatures do not verify.
	ErrLicenseKeyInvalid = NewErr(ErrPreconditionFailed, "ERR_LICENSE_KEY_INVALID")
	// ErrLicenseKeyOtherInstallation is a key for another server. Its param InstallationID is this one's.
	ErrLicenseKeyOtherInstallation = NewErr(ErrPreconditionFailed, "ERR_LICENSE_KEY_OTHER_INSTALLATION")
	// ErrLicenseKeyExpired is a key past its grace.
	ErrLicenseKeyExpired = NewErr(ErrPreconditionFailed, "ERR_LICENSE_KEY_EXPIRED")
)
