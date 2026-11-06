package base

// AuditLogType is what happened, named after the action rather than after the
// endpoint it arrived on.
//
// An endpoint-shaped name (say "get-reveal") ties recorded history to the routing
// of the day it was written: move the action to another route later and the same
// event starts being recorded under a different name, so any question asked across
// that boundary quietly gets a wrong answer. Where the action came in is recorded
// separately, as AuditLogSource.
type AuditLogType string

const (
	// AuditLogTypeSecretReveal records a stored secret being handed out in the
	// clear, whether or not the caller was allowed to have it.
	AuditLogTypeSecretReveal AuditLogType = "secret-reveal"

	// AuditLogTypeSettingCreate, AuditLogTypeSettingUpdate,
	// AuditLogTypeSettingStatusUpdate and AuditLogTypeSettingDelete record the
	// lifecycle of a stored setting - a registry credential, an SSH key, a backup
	// repository, and the twenty-odd others.
	//
	// Four types for all of them, not four per kind. Which kind it was is already
	// on the entry as ResType, and folding it into the name instead would put
	// eighty-odd values in the type filter, which is the same as having no filter.
	//
	// Creation is recorded alongside the rest because a trail that shows a setting
	// changed and then deleted, with no record of it ever appearing, reads as if
	// something is missing from the log.
	AuditLogTypeSettingCreate AuditLogType = "setting-create"
	AuditLogTypeSettingUpdate AuditLogType = "setting-update"

	// AuditLogTypeSettingStatusUpdate is separate from a plain update because it
	// covers the switches that decide whether a setting is in force at all -
	// status, expiry, inheritable, default - and "who disabled this" is a question
	// asked on its own.
	AuditLogTypeSettingStatusUpdate AuditLogType = "setting-status-update"

	AuditLogTypeSettingDelete AuditLogType = "setting-delete"

	// AuditLogTypeTaskCancel records work in flight being stopped by hand: a
	// deploy, an image build, a backup, a cleanup sweep.
	//
	// The only type tasks get, and it is about the person rather than the task.
	// Everything else in a task's life - queued, retried, failed, finished - is
	// the system acting on its own and already has the task row as its record.
	// A cancel is the one point where somebody reaches in and stops the work,
	// which is what is worth answering for later.
	//
	// One type across every scope, rather than folding the cancel into whatever
	// the task was doing. Tasks run under projects, environments, apps, users and
	// the install itself, and there is no type at all for some of those - the
	// entry's own scope says where it happened, which is what a type per scope
	// would have said at the cost of five more values in the type filter.
	AuditLogTypeTaskCancel AuditLogType = "task-cancel"

	// AuditLogTypeUserCreate, AuditLogTypeUserUpdate and AuditLogTypeUserDelete
	// record an account appearing, being changed, and being taken away.
	//
	// Split three ways for the reason projects and apps are: an account arriving
	// and an account disappearing are what somebody scanning the type column is
	// looking for, and folded into the edits a new admin would be
	// indistinguishable from a changed job title.
	//
	// One type for every edit, with section naming which form was used - account,
	// profile, password, password-reset, password-reset-request, mfa-setup,
	// mfa-remove. The two weightiest, a role change and a second factor being
	// taken off, are sections on the same reasoning as everything else in this
	// file; what moved is in the entry's detail, with the values, because a role
	// is not a secret and "member to admin" is the whole of what a reader wants.
	AuditLogTypeUserCreate AuditLogType = "user-create"
	AuditLogTypeUserUpdate AuditLogType = "user-update"
	AuditLogTypeUserDelete AuditLogType = "user-delete"

	// AuditLogTypeUserLogin records a session being handed out, and every attempt
	// that was turned away.
	//
	// The refusals are the half that exists nowhere else. A login that worked
	// leaves a session behind to find; a wrong password leaves nothing at all,
	// and a run of them against one account is the earliest notice anybody gets
	// that somebody is working at it. The backoff that slows those guesses down
	// counts them in redis, where the count is per account, expires in hours, and
	// is never read by a person.
	//
	// One type for every way in, with section naming which - password, passcode,
	// oauth, dev-mode. A type per method would put four values in the
	// filter to answer one question, and which door was used is the second
	// question a reader asks, not the first.
	//
	// A refused attempt names no actor. Who was on the other end is precisely
	// what could not be established; the account that was aimed at is recorded as
	// the entry's resource, and where the attempt came from is on every entry
	// anyway.
	AuditLogTypeUserLogin AuditLogType = "user-login"

	// AuditLogTypeUserLogout records a session being given up - the one in hand,
	// or every session the user has, which section tells apart.
	//
	// Worth recording for the same reason a login is: signing out everywhere is
	// what somebody does after losing a laptop, and also what somebody else does
	// after taking one. The two read identically in the moment and differently
	// afterwards, which is what a trail is for.
	AuditLogTypeUserLogout AuditLogType = "user-logout"

	// AuditLogTypeLicenseUpdate records a license key being installed or removed, with section naming which.
	AuditLogTypeLicenseUpdate AuditLogType = "license-update"

	// AuditLogTypeSSHCAUpdate records a change to who may get an SSH certificate and how: a CA made or deleted, the
	// lifetime of certificates, or a person's principals, with section naming which.
	AuditLogTypeSSHCAUpdate AuditLogType = "ssh-ca-update"
	// AuditLogTypeSSHCertIssue records an SSH certificate being signed. Certificates are not stored: this is the record
	// of who holds which, and when each ends.
	AuditLogTypeSSHCertIssue AuditLogType = "ssh-cert-issue"
	// AuditLogTypeOrgMemberUpdate records a member's role or status in an organization changing, with section naming
	// which, and what it was and became.
	AuditLogTypeOrgMemberUpdate AuditLogType = "org-member-update"
	// AuditLogTypeOrgGroupUpdate records a group made, changed or deleted, or its members set, with section naming which.
	AuditLogTypeOrgGroupUpdate AuditLogType = "org-group-update"
	// AuditLogTypeUserIdentityUpdate records a person publishing their identity, or another in its place, with its
	// fingerprint.
	AuditLogTypeUserIdentityUpdate AuditLogType = "user-identity-update"
)

var AllAuditLogTypes = []AuditLogType{
	AuditLogTypeLicenseUpdate,
	AuditLogTypeSSHCAUpdate,
	AuditLogTypeSSHCertIssue,
	AuditLogTypeOrgMemberUpdate,
	AuditLogTypeOrgGroupUpdate,
	AuditLogTypeUserIdentityUpdate,
	AuditLogTypeSecretReveal,
	AuditLogTypeUserCreate,
	AuditLogTypeUserUpdate,
	AuditLogTypeUserDelete,
	AuditLogTypeUserLogin,
	AuditLogTypeUserLogout,
	AuditLogTypeSettingCreate,
	AuditLogTypeSettingUpdate,
	AuditLogTypeSettingStatusUpdate,
	AuditLogTypeSettingDelete,
	AuditLogTypeTaskCancel,
}

// AuditLogSource is the way in - which endpoint, or which subsystem.
type AuditLogSource string

const (
	// AuditLogSourceAPIGet is a plain GET asked to include the secrets.
	AuditLogSourceAPIGet AuditLogSource = "api-get"

	// AuditLogSourceAPICreate is a create endpoint.
	AuditLogSourceAPICreate AuditLogSource = "api-create"

	// AuditLogSourceAPIUpdate is an update endpoint.
	AuditLogSourceAPIUpdate AuditLogSource = "api-update"

	// AuditLogSourceAPIDelete is a delete endpoint.
	AuditLogSourceAPIDelete AuditLogSource = "api-delete"

	// AuditLogSourceAPIAction is an endpoint that makes something happen rather
	// than writing a record: a deploy, a restart, a stop, a cancel. Recording one
	// of those as api-update would say the stored configuration changed, which is
	// the opposite of what a reader needs to know about a restart.
	AuditLogSourceAPIAction AuditLogSource = "api-action"
)

// AuditLogResult says whether the action was permitted.
//
// The refusals are the more interesting half: a denied attempt is the only signal
// that somebody is trying doors, and it is invisible if only successes are kept.
type AuditLogResult string

const (
	AuditLogResultAllowed AuditLogResult = "allowed"
	AuditLogResultDenied  AuditLogResult = "denied"
)
