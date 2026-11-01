package cacheentity

import "github.com/hivepaas/shellwing-server/base"

type TaskControl struct {
	ID  string           `json:"id"`
	Cmd base.TaskCommand `json:"cmd"`
}
