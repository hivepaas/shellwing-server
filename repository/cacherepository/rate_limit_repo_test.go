package cacherepository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/shellwing-server/pkg/ulid"
)

// This needs a real Redis: what it checks is that the counter and its expiry are one step, and that the
// expiry is set once per window, which no fake reproduces. Point SHW_TEST_REDIS_URL at one, such as
// redis://localhost:6379/15, to enable it.
func TestRateLimitHitCountsInAFixedWindow(t *testing.T) {
	url := os.Getenv("SHW_TEST_REDIS_URL")
	if url == "" {
		t.Skip("SHW_TEST_REDIS_URL not set, skipping the rate limit test")
	}
	options, err := redis.ParseURL(url)
	require.NoError(t, err)
	client := redis.NewClient(options)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	repo := NewRateLimitRepo(client)
	key := "test:" + gofn.Must(ulid.NewStringULID())
	t.Cleanup(func() { _ = client.Del(ctx, "rate-limit:"+key).Err() })
	const window = 2 * time.Second

	for want := int64(1); want <= 3; want++ {
		got, err := repo.Hit(ctx, key, window)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	}

	// The window starts with its first request: later requests do not push its end back.
	ttl, err := client.PTTL(ctx, "rate-limit:"+key).Result()
	require.NoError(t, err)
	assert.Positive(t, ttl)
	assert.LessOrEqual(t, ttl, window)
	time.Sleep(window / 2)
	_, err = repo.Hit(ctx, key, window)
	require.NoError(t, err)
	ttlAfter, err := client.PTTL(ctx, "rate-limit:"+key).Result()
	require.NoError(t, err)
	assert.Less(t, ttlAfter, ttl)

	// A new window counts from one again.
	time.Sleep(ttlAfter + 100*time.Millisecond)
	got, err := repo.Hit(ctx, key, window)
	require.NoError(t, err)
	assert.Equal(t, int64(1), got)
}
