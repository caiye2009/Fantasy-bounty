package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

const requestContextKey = "request_context"

// RequestContext holds unified request metadata populated by middleware and handlers.
type RequestContext struct {
	RequestID  string
	ClientIP   string
	UserAgent  string
	StartTime  time.Time
	UserID     string         // filled by Auth middleware
	Username   string         // filled by Auth middleware
	Action     string         // set by handler (optional)
	Resource   string         // set by handler (optional)
	ResourceID string         // set by handler (optional)
	Detail     map[string]any // set by handler (optional)
}

// SetRequestContext stores the RequestContext in the gin context.
// Should only be called by middleware.
func SetRequestContext(c *gin.Context, rc *RequestContext) {
	c.Set(requestContextKey, rc)
}

// GetRequestContext retrieves the RequestContext from the gin context.
// Returns nil if not set.
func GetRequestContext(c *gin.Context) *RequestContext {
	v, exists := c.Get(requestContextKey)
	if !exists {
		return nil
	}
	rc, ok := v.(*RequestContext)
	if !ok {
		return nil
	}
	return rc
}
