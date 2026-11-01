package reslinkserviceimpl

import (
	"github.com/hivepaas/shellwing-server/repository"
	"github.com/hivepaas/shellwing-server/service/reslinkservice"
)

func New(
	resLinkRepo repository.ResLinkRepo,
) reslinkservice.Service {
	return &service{
		resLinkRepo: resLinkRepo,
	}
}

type service struct {
	resLinkRepo repository.ResLinkRepo
}
