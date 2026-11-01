package cacherepository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/rediscache"
)

type RateLimitRepo interface {
	// Hit counts one request against key in a fixed window that starts with the window's first request,
	// and returns the count so far, this one included. The caller compares it with its limit; the key
	// names both what is limited and by what, such as "license-refresh:ip:203.0.113.7".
	Hit(ctx context.Context, key string, window time.Duration) (int64, error)
}

type rateLimitRepo struct {
	client rediscache.Client
}

func NewRateLimitRepo(client rediscache.Client) RateLimitRepo {
	return &rateLimitRepo{client: client}
}

func (repo *rateLimitRepo) Hit(ctx context.Context, key string, window time.Duration) (int64, error) {
	var count *redis.IntCmd
	// One transaction, so a counter is never left without its expiry by a failure between the two. NX
	// sets the expiry on the window's first request only: the window does not slide with every request,
	// which would keep a steady caller locked out for good.
	_, err := repo.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		count = p.Incr(ctx, repo.formatKey(key))
		p.ExpireNX(ctx, repo.formatKey(key), window)
		return nil
	})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	return count.Val(), nil
}

func (repo *rateLimitRepo) formatKey(key string) string {
	return "rate-limit:" + key
}
