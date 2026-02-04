package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestContextMiddleware creates a new RequestContext for every incoming request.
func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rc := &RequestContext{
			RequestID: uuid.New().String(),
			ClientIP:  c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			StartTime: time.Now(),
		}
		SetRequestContext(c, rc)
		c.Next()
	}
}
