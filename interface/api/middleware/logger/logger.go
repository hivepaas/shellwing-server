package logger

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/shellwing-server/pkg/logging"
)

func Logger(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(logging.LoggerCtxKey, logger)

		c.Next()
	}
}
