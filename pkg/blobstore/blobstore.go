// Package blobstore keeps the content of synced files, in a local folder or in S3-compatible storage. It knows nothing
// of what it keeps: a key names some bytes, and whoever writes under a key never changes what is there.
package blobstore

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrNotFound is a key the store has nothing under.
var ErrNotFound = errors.New("blob not found")

// ErrKeyInvalid is a key that could reach outside the store.
var ErrKeyInvalid = errors.New("blob key is invalid")

// Blob is one entry of a listing.
type Blob struct {
	Key      string
	Size     int64
	Modified time.Time
}

type Store interface {
	// Check says whether the store can be used, so that a bad setting stops the start rather than the first sync.
	Check(ctx context.Context) error
	// Put writes data under key whole, or not at all: a reader never finds half of it.
	Put(ctx context.Context, key string, data []byte) error
	// Get reads what is under key, or ErrNotFound.
	Get(ctx context.Context, key string) ([]byte, error)
	// Delete removes what is under key. Nothing there is no error.
	Delete(ctx context.Context, key string) error
	// List calls fn for each key starting with prefix, in no promised order but with keys of the same folder together.
	List(ctx context.Context, prefix string, fn func(Blob) error) error
}

func validKey(key string) bool {
	if key == "" || strings.HasPrefix(key, "/") || strings.Contains(key, "\\") {
		return false
	}
	for segment := range strings.SplitSeq(key, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}
