package blobstore

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hivepaas/shellwing-server/hperrors"
)

const (
	localDirPerm  = 0o700
	localFilePerm = 0o600
	// Half-written files start with this. No key has a segment that does: see tempName.
	localTempPrefix = ".tmp-"
)

type localStore struct {
	root string
}

// NewLocal keeps blobs in the folder root. It suits one instance: a second needs the same folder, and what a network
// filesystem promises about renames and syncing is not something this relies on.
func NewLocal(root string) (Store, error) {
	if root == "" {
		return nil, fmt.Errorf("%w: the folder is not set", ErrKeyInvalid)
	}
	return &localStore{root: root}, nil
}

func (s *localStore) Check(_ context.Context) error {
	if err := os.MkdirAll(s.root, localDirPerm); err != nil {
		return hperrors.Wrap(err)
	}
	tmp, err := os.CreateTemp(s.root, localTempPrefix+"*")
	if err != nil {
		return hperrors.Wrap(err)
	}
	_ = tmp.Close()
	return hperrors.Wrap(os.Remove(tmp.Name()))
}

func (s *localStore) path(key string) (string, error) {
	if !validKey(key) {
		return "", fmt.Errorf("%w: %q", ErrKeyInvalid, key)
	}
	return filepath.Join(s.root, filepath.FromSlash(key)), nil
}

func (s *localStore) Put(_ context.Context, key string, data []byte) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, localDirPerm); err != nil {
		return hperrors.Wrap(err)
	}
	tmp, err := os.CreateTemp(dir, localTempPrefix+"*")
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err = tmp.Chmod(localFilePerm); err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		// What a client was told is written has to outlive a crash of the machine.
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = os.Rename(tmp.Name(), path); err != nil {
		return hperrors.Wrap(err)
	}
	return syncDir(dir)
}

func syncDir(path string) error {
	dir, err := os.Open(path) //nolint:gosec // G304: a path made of a checked key
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() { _ = dir.Close() }()
	return hperrors.Wrap(dir.Sync())
}

func (s *localStore) Get(_ context.Context, key string) ([]byte, error) {
	path, err := s.path(key)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path) //nolint:gosec // G304: a path made of a checked key
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return data, nil
}

func (s *localStore) Delete(_ context.Context, key string) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return hperrors.Wrap(err)
	}
	s.removeEmptyParents(filepath.Dir(path))
	return nil
}

// removeEmptyParents takes away the folders a delete left empty. A record's versions each have a folder of their own
// otherwise forever. os.Remove refuses a folder with something in it, so a write under way loses at worst its folder,
// and Put makes it again.
func (s *localStore) removeEmptyParents(dir string) {
	for dir != s.root && strings.HasPrefix(dir, s.root) {
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

func (s *localStore) List(ctx context.Context, prefix string, fn func(Blob) error) error {
	err := filepath.WalkDir(s.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			// A folder a delete just removed.
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return hperrors.Wrap(err)
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return hperrors.Wrap(ctxErr)
		}
		if entry.IsDir() || strings.HasPrefix(entry.Name(), localTempPrefix) {
			return nil
		}
		rel, err := filepath.Rel(s.root, path)
		if err != nil {
			return hperrors.Wrap(err)
		}
		key := filepath.ToSlash(rel)
		if !strings.HasPrefix(key, prefix) {
			return nil
		}
		info, err := entry.Info()
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return hperrors.Wrap(err)
		}
		return fn(Blob{Key: key, Size: info.Size(), Modified: info.ModTime()})
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return hperrors.Wrap(err)
}

// Root is the folder the store keeps blobs in.
func (s *localStore) Root() string {
	return s.root
}
