package base

// FreeUserLimit is how many users a server may have with no license in force.
const FreeUserLimit = 2

// NoUserLimit is the limit of a key that sets none: a limit left out of a key is no limit.
const NoUserLimit = -1

type LicenseStatus string

const (
	// LicenseStatusNone: no key is installed.
	LicenseStatusNone  LicenseStatus = "none"
	LicenseStatusValid LicenseStatus = "valid"
	LicenseStatusTrial LicenseStatus = "trial"
	// LicenseStatusGrace: past its expiry, and still in force for its grace.
	LicenseStatusGrace LicenseStatus = "grace"
	// LicenseStatusExpired: past its grace. The free limit is back, which only stops creating.
	LicenseStatusExpired LicenseStatus = "expired"
	// LicenseStatusInvalid: the installed key no longer verifies, as under another set of public keys.
	LicenseStatusInvalid LicenseStatus = "invalid"
	// LicenseStatusOtherInstallation: the key names another server.
	LicenseStatusOtherInstallation LicenseStatus = "other-installation"
)
