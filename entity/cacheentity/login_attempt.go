package cacheentity

import "github.com/hivepaas/shellwing-server/pkg/failbackoff"

// LoginAttempt counts one user's consecutive failed password checks.
//
// Its own key space, separate from [AppSecretAttempt]: failing to log in and
// failing to prove you are the operator are different runs of failures and must
// not shorten each other's fuse.
type LoginAttempt = failbackoff.Attempt
