package middleware

import (
	"net/http"

	"back/pkg/internal_token"

	"github.com/gin-gonic/gin"
)

const internalTokenKey = "internal_token"

// InternalToken fetches a global internal-system token and stores it in gin context.
func InternalToken(tokenManager *internal_token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		t, err := tokenManager.GetToken()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    http.StatusServiceUnavailable,
				"message": "failed to get internal token",
			})
			c.Abort()
			return
		}
		c.Set(internalTokenKey, t)
		c.Next()
	}
}

func GetInternalToken(c *gin.Context) (string, bool) {
	v, ok := c.Get(internalTokenKey)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

