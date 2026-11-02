package syncserviceimpl

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/hivepaas/shellwing-server/entity"
	"github.com/hivepaas/shellwing-server/hperrors"
)

const (
	recordsPrefix  = "records/"
	recordSuffix   = ".sw"
	kindHost       = "host"
	kindMaxLen     = 20
	nameFileMaxLen = 200
)

// The layout clients sync, and nothing else: names become keys in the store.
var objectNameRe = regexp.MustCompile(`^(?:format\.json|(?:records|attachments|pairing)/[A-Za-z0-9][A-Za-z0-9._-]*)$`)

func validateName(name string) error {
	if len(name) > len(recordsPrefix)+nameFileMaxLen || !objectNameRe.MatchString(name) {
		return hperrors.Wrap(hperrors.ErrSyncObjectNameInvalid).WithMsgLog("name %q", name)
	}
	return nil
}

func etagOf(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// clearPart is what a record shows without its key. The server counts by it, and never needs more.
type clearPart struct {
	ID   string  `json:"id"`
	Kind string  `json:"kind"`
	Data *string `json:"data"`
}

// describe fills what the line says about a file from its content: a record's clear kind, and whether it is a
// tombstone. Other files have neither.
func describe(object *entity.SyncObject, data []byte) error {
	object.Kind, object.Tombstone = "", false
	if !strings.HasPrefix(object.Name, recordsPrefix) {
		return nil
	}
	var clear clearPart
	if err := json.Unmarshal(data, &clear); err != nil {
		return hperrors.Wrap(hperrors.ErrSyncRecordInvalid).WithCause(err)
	}
	if clear.Kind == "" || len(clear.Kind) > kindMaxLen || object.Name != recordsPrefix+clear.ID+recordSuffix {
		return hperrors.Wrap(hperrors.ErrSyncRecordInvalid).WithMsgLog("record %q, kind %q", clear.ID, clear.Kind)
	}
	object.Kind = clear.Kind
	object.Tombstone = clear.Data == nil
	return nil
}

// counts adds what one line weighs in a space's totals, with sign 1 or -1.
func counts(space *entity.SyncSpace, object *entity.SyncObject, sign int) {
	if !object.IsLive() {
		return
	}
	space.BytesUsed += int64(sign) * object.Size
	if object.Tombstone || !strings.HasPrefix(object.Name, recordsPrefix) {
		return
	}
	space.RecordCount += sign
	if object.Kind == kindHost {
		space.HostCount += sign
	}
}
