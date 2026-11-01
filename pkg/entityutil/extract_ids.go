package entityutil

import (
	"github.com/hivepaas/shellwing-server/entity"
)

func ExtractIDs[T entity.IDEntity, S ~[]T](entities S) []string {
	result := make([]string, 0, len(entities))
	for _, ent := range entities {
		result = append(result, ent.GetID())
	}
	return result
}
