package cacherepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/hivepaas/shellwing-server/entity/cacheentity"
	"github.com/hivepaas/shellwing-server/hperrors"
	"github.com/hivepaas/shellwing-server/infra/rediscache"
)

type DeviceAuthCodeRepo interface {
	Set(ctx context.Context, codeHash string, code *cacheentity.DeviceAuthCode, exp time.Duration) error
	// Take returns the code and removes it: it works once. It returns nil when there is none.
	Take(ctx context.Context, codeHash string) (*cacheentity.DeviceAuthCode, error)
}

type deviceAuthCodeRepo struct {
	client rediscache.Client
}

func NewDeviceAuthCodeRepo(client rediscache.Client) DeviceAuthCodeRepo {
	return &deviceAuthCodeRepo{client: client}
}

func (repo *deviceAuthCodeRepo) Set(ctx context.Context, codeHash string, code *cacheentity.DeviceAuthCode,
	exp time.Duration) error {
	data, err := json.Marshal(code)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(repo.client.Set(ctx, repo.formatKey(codeHash), data, exp).Err())
}

func (repo *deviceAuthCodeRepo) Take(ctx context.Context, codeHash string) (*cacheentity.DeviceAuthCode, error) {
	data, err := repo.client.GetDel(ctx, repo.formatKey(codeHash)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	code := &cacheentity.DeviceAuthCode{}
	if err = json.Unmarshal(data, code); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return code, nil
}

func (repo *deviceAuthCodeRepo) formatKey(codeHash string) string {
	return fmt.Sprintf("device-auth-code:%s", codeHash)
}
