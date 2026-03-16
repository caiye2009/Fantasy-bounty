package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole enforces that the JWT-authenticated user has the given role.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rc := GetRequestContext(c)
		if rc == nil || rc.Role == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "missing role in context",
			})
			c.Abort()
			return
		}
		if rc.Role != role {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "forbidden",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

