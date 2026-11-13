package datakeyserviceimpl

import (
	"github.com/hivepaas/shellwing-server/pkg/logging"
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/datakeyservice"
)

type service struct {
	encryptionKeyRepo repository.EncryptionKeyRepo
	logger            logging.Logger
}

func New(
	encryptionKeyRepo repository.EncryptionKeyRepo,
	logger logging.Logger,
) datakeyservice.Service {
	return &service{
		encryptionKeyRepo: encryptionKeyRepo,
		logger:            logger,
	}
}
