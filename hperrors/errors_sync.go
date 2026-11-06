package hperrors

// Errors for syncing files.
var (
	// ErrSyncObjectChanged refuses a conditional write: the file is not as the writer last saw it. It reads the file
	// again, merges, and writes once more.
	ErrSyncObjectChanged = NewErr(ErrPreconditionFailed, "ERR_SYNC_OBJECT_CHANGED")
	// ErrSyncObjectNameInvalid is a name outside the layout clients sync.
	ErrSyncObjectNameInvalid = NewErr(ErrArgumentInvalid, "ERR_SYNC_OBJECT_NAME_INVALID")
	// ErrSyncRecordInvalid is a record whose clear part cannot be read, or names another record.
	ErrSyncRecordInvalid = NewErr(ErrArgumentInvalid, "ERR_SYNC_RECORD_INVALID")
	// ErrSyncObjectTooLarge is a file above the server's limit for one.
	ErrSyncObjectTooLarge = NewErr(ErrArgumentInvalid, "ERR_SYNC_OBJECT_TOO_LARGE")
	// ErrSyncSpaceFull refuses a write that would take the space past its limit.
	ErrSyncSpaceFull = NewErr(ErrConflict, "ERR_SYNC_SPACE_FULL")

	// ErrDeviceRedirectInvalid is an address that is no app's loopback listener: a code never goes anywhere else.
	ErrDeviceRedirectInvalid = NewErr(ErrArgumentInvalid, "ERR_DEVICE_REDIRECT_INVALID")
	// ErrDeviceCodeInvalid is a code that is unknown, used already, too old, or not this app's. An app signs in again.
	ErrDeviceCodeInvalid = NewErr(ErrArgumentInvalid, "ERR_DEVICE_CODE_INVALID")
	// ErrDeviceTokenInvalid is a refresh token no device may use: its person signs in again.
	ErrDeviceTokenInvalid = NewErr(ErrUnauthorized, "ERR_DEVICE_TOKEN_INVALID")
	// ErrScopeNotAllowed refuses a token narrowed to sync anywhere else.
	ErrScopeNotAllowed = NewErr(ErrForbidden, "ERR_SCOPE_NOT_ALLOWED")
)

// Errors for SSH certificates. A rule that refuses is 412, as for the limit of users: none is about who is asking.
var (
	// ErrSSHCANotSetUp is an organization with no active CA: an admin makes one.
	ErrSSHCANotSetUp = NewErr(ErrPreconditionFailed, "ERR_SSH_CA_NOT_SET_UP")
	// ErrSSHCAActive refuses deleting the CA that signs: making a new one retires it.
	ErrSSHCAActive = NewErr(ErrPreconditionFailed, "ERR_SSH_CA_ACTIVE")
	// ErrSSHNoPrincipals is a person with no account a certificate could log in as: an admin grants them.
	ErrSSHNoPrincipals = NewErr(ErrPreconditionFailed, "ERR_SSH_NO_PRINCIPALS")
	// ErrSSHPublicKeyInvalid is a text that is no public key, or a kind or size of one that is not signed.
	ErrSSHPublicKeyInvalid = NewErr(ErrArgumentInvalid, "ERR_SSH_PUBLIC_KEY_INVALID")
	// ErrSSHPrincipalInvalid is a name no account could have, or that would mean something else in a server's
	// configuration. Its param Name is the name.
	ErrSSHPrincipalInvalid = NewErr(ErrArgumentInvalid, "ERR_SSH_PRINCIPAL_INVALID")
)

// Errors for organizations' members.
var (
	// ErrOrgLastOwner refuses a change that would leave the organization with no active owner: another is made first.
	ErrOrgLastOwner = NewErr(ErrPreconditionFailed, "ERR_ORG_LAST_OWNER")
)

// Errors for people's identities.
var (
	// ErrIdentityInvalid is an identity whose keys are not the right size, or whose X25519 key its Ed25519 key did not
	// sign.
	ErrIdentityInvalid = NewErr(ErrArgumentInvalid, "ERR_IDENTITY_INVALID")
)

// Errors for collections and their keys.
var (
	// ErrIdentityMissing is a person with no identity published: their app makes one when it syncs with this server.
	ErrIdentityMissing = NewErr(ErrPreconditionFailed, "ERR_IDENTITY_MISSING")
	// ErrCollectionReadOnly refuses a write from somebody who may only connect.
	ErrCollectionReadOnly = NewErr(ErrForbidden, "ERR_COLLECTION_READ_ONLY")
	// ErrCollectionKeyVersion is a key of a version other than the one expected: another app made one first. The app
	// fetches the collection again.
	ErrCollectionKeyVersion = NewErr(ErrConflict, "ERR_COLLECTION_KEY_VERSION")
	// ErrCollectionNoKey is a person who does not hold the collection's current key, doing what needs it.
	ErrCollectionNoKey = NewErr(ErrPreconditionFailed, "ERR_COLLECTION_NO_KEY")
	// ErrCollectionWrapInvalid is a wrap not signed by the caller's identity, of the wrong size, for somebody without
	// access, for another key than their identity's, or for somebody who has it already. Its param Name is whom it is
	// for.
	ErrCollectionWrapInvalid = NewErr(ErrArgumentInvalid, "ERR_COLLECTION_WRAP_INVALID")
	// ErrCollectionWrapsMissing is a new key not wrapped for the caller and every owner with an identity. Its param
	// Name is whom it lacks.
	ErrCollectionWrapsMissing = NewErr(ErrPreconditionFailed, "ERR_COLLECTION_WRAPS_MISSING")
)
