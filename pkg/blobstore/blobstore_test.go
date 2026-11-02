package blobstore

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hivepaas/shellwing-server/pkg/ulid"
)

// What every store has to do. The S3 part needs a real server: point SHW_TEST_S3_ENDPOINT, _BUCKET, _ACCESS_KEY and
// _SECRET_KEY at one, whose bucket exists, to enable it.
func storesUnderTest(t *testing.T) map[string]Store {
	t.Helper()
	local, err := NewLocal(filepath.Join(t.TempDir(), "blobs"))
	require.NoError(t, err)
	stores := map[string]Store{"local": local}

	if endpoint := os.Getenv("SHW_TEST_S3_ENDPOINT"); endpoint != "" {
		id, err := ulid.NewStringULID()
		require.NoError(t, err)
		remote, err := NewS3(&S3Config{
			Endpoint:  endpoint,
			Region:    "us-east-1",
			Bucket:    os.Getenv("SHW_TEST_S3_BUCKET"),
			Prefix:    "test-" + id,
			AccessKey: os.Getenv("SHW_TEST_S3_ACCESS_KEY"),
			SecretKey: os.Getenv("SHW_TEST_S3_SECRET_KEY"),
			PathStyle: true,
		})
		require.NoError(t, err)
		stores["s3"] = remote
	}
	return stores
}

func TestStores(t *testing.T) {
	for name, store := range storesUnderTest(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			require.NoError(t, store.Check(ctx))

			_, err := store.Get(ctx, "spaces/a/records/x.sw/none")
			require.ErrorIs(t, err, ErrNotFound)
			require.NoError(t, store.Delete(ctx, "spaces/a/records/x.sw/none"), "nothing there is no error")

			content := bytes.Repeat([]byte("content\x00"), 1000)
			require.NoError(t, store.Put(ctx, "spaces/a/records/x.sw/e1", content))
			require.NoError(t, store.Put(ctx, "spaces/a/format.json/e2", []byte("{}")))
			require.NoError(t, store.Put(ctx, "spaces/b/records/y.sw/e3", nil), "an empty file is a file")

			got, err := store.Get(ctx, "spaces/a/records/x.sw/e1")
			require.NoError(t, err)
			assert.True(t, bytes.Equal(content, got))
			got, err = store.Get(ctx, "spaces/b/records/y.sw/e3")
			require.NoError(t, err)
			assert.Empty(t, got)

			var inA []string
			require.NoError(t, store.List(ctx, "spaces/a/", func(b Blob) error {
				inA = append(inA, b.Key)
				assert.False(t, b.Modified.IsZero())
				if b.Key == "spaces/a/records/x.sw/e1" {
					assert.Equal(t, int64(len(content)), b.Size)
				}
				return nil
			}))
			slices.Sort(inA)
			assert.Equal(t, []string{"spaces/a/format.json/e2", "spaces/a/records/x.sw/e1"}, inA)

			require.NoError(t, store.Delete(ctx, "spaces/a/records/x.sw/e1"))
			_, err = store.Get(ctx, "spaces/a/records/x.sw/e1")
			require.ErrorIs(t, err, ErrNotFound)
			got, err = store.Get(ctx, "spaces/a/format.json/e2")
			require.NoError(t, err)
			assert.Equal(t, []byte("{}"), got, "the file beside it stays")

			require.NoError(t, store.Delete(ctx, "spaces/a/format.json/e2"))
			require.NoError(t, store.Delete(ctx, "spaces/b/records/y.sw/e3"))
		})
	}
}

func TestStoresRefuseKeysThatLeaveThem(t *testing.T) {
	for name, store := range storesUnderTest(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			for _, key := range []string{"", "/etc/passwd", "../x", "a/../../x", "a//b", "a/./b", `a\b`} {
				assert.ErrorIs(t, store.Put(ctx, key, []byte("x")), ErrKeyInvalid, key)
				_, err := store.Get(ctx, key)
				assert.ErrorIs(t, err, ErrKeyInvalid, key)
				assert.ErrorIs(t, store.Delete(ctx, key), ErrKeyInvalid, key)
			}
		})
	}
}

func TestLocalStoreLeavesNoEmptyFoldersOrTempFiles(t *testing.T) {
	root := filepath.Join(t.TempDir(), "blobs")
	store, err := NewLocal(root)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, store.Check(ctx))

	require.NoError(t, store.Put(ctx, "spaces/a/records/x.sw/e1", []byte("x")))
	info, err := os.Stat(filepath.Join(root, "spaces/a/records/x.sw/e1"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(localFilePerm), info.Mode().Perm())
	entries, err := os.ReadDir(filepath.Join(root, "spaces/a/records/x.sw"))
	require.NoError(t, err)
	assert.Len(t, entries, 1, "no temp file left behind")

	require.NoError(t, store.Delete(ctx, "spaces/a/records/x.sw/e1"))
	entries, err = os.ReadDir(root)
	require.NoError(t, err)
	assert.Empty(t, entries)
}
