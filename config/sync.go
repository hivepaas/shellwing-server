package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"
)

const (
	SyncStoreLocal = "local"
	SyncStoreS3    = "s3"
)

var ErrSyncConfigInvalid = errors.New("sync config is invalid")

// Sync is where synced files are kept, and how much a space may hold.
type Sync struct {
	// MaxObjectBytes is the largest file a client may sync. The largest it makes is an image for a note: 2 MB, sealed
	// and in base64.
	MaxObjectBytes int64 `toml:"max_object_bytes" env:"SHW_SYNC_MAX_OBJECT_BYTES" default:"8388608"`
	// MaxSpaceBytes is what one space may hold in all. Zero: no limit.
	MaxSpaceBytes int64 `toml:"max_space_bytes" env:"SHW_SYNC_MAX_SPACE_BYTES" default:"0"`
	// TrashDelay is how long replaced content stays, for reads already under way.
	TrashDelay time.Duration `toml:"trash_delay" env:"SHW_SYNC_TRASH_DELAY" default:"1h"`
	// CleanInterval is how often that content is removed.
	CleanInterval time.Duration `toml:"clean_interval" env:"SHW_SYNC_CLEAN_INTERVAL" default:"10m"`

	Store SyncStore `toml:"store"`
}

// SyncStore is the store itself. A local folder suits one instance; several need S3.
type SyncStore struct {
	Type string `toml:"type" env:"SHW_SYNC_STORE_TYPE" default:"local"`
	// Path is the folder of the local store. Empty: "sync" in the app path.
	Path string `toml:"path" env:"SHW_SYNC_STORE_PATH"`

	Endpoint  string `toml:"endpoint" env:"SHW_SYNC_STORE_ENDPOINT"`
	Region    string `toml:"region" env:"SHW_SYNC_STORE_REGION" default:"auto"`
	Bucket    string `toml:"bucket" env:"SHW_SYNC_STORE_BUCKET"`
	Prefix    string `toml:"prefix" env:"SHW_SYNC_STORE_PREFIX"`
	AccessKey string `toml:"access_key" env:"SHW_SYNC_STORE_ACCESS_KEY"`
	SecretKey string `toml:"secret_key" env:"SHW_SYNC_STORE_SECRET_KEY"`
	PathStyle bool   `toml:"path_style" env:"SHW_SYNC_STORE_PATH_STYLE"`
}

// SyncStorePath is the folder of the local store.
func (cfg *Config) SyncStorePath() string {
	if cfg.Sync.Store.Path != "" {
		return cfg.Sync.Store.Path
	}
	return filepath.Join(cfg.AppPath, "sync")
}

func (s *Sync) validate() error {
	switch s.Store.Type {
	case SyncStoreLocal:
	case SyncStoreS3:
		if s.Store.Bucket == "" {
			return fmt.Errorf("%w: SHW_SYNC_STORE_TYPE is s3, and SHW_SYNC_STORE_BUCKET is not set", ErrSyncConfigInvalid)
		}
	default:
		return fmt.Errorf("%w: SHW_SYNC_STORE_TYPE %q: local or s3", ErrSyncConfigInvalid, s.Store.Type)
	}
	if s.MaxObjectBytes <= 0 {
		return fmt.Errorf("%w: SHW_SYNC_MAX_OBJECT_BYTES must be above zero", ErrSyncConfigInvalid)
	}
	return nil
}
