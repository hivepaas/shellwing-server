package datakeyserviceimpl

import (
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/datakeyservice"
)

type service struct {
	encryptionKeyRepo repository.EncryptionKeyRepo
}

func New(
	encryptionKeyRepo repository.EncryptionKeyRepo,
) datakeyservice.Service {
	return &service{
		encryptionKeyRepo: encryptionKeyRepo,
	}
}
