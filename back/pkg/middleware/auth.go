package middleware

import (
	"back/pkg/jwt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth(jwtService *jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		log.Printf("[JWT-AUTH] 收到请求: %s %s", c.Request.Method, path)

		// 从Header中获取Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("[JWT-AUTH] 缺少 Authorization header, 路径: %s", path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "missing authorization header",
			})
			c.Abort()
			return
		}

		// 截断输出token
		maskedAuth := authHeader
		if len(maskedAuth) > 25 {
			maskedAuth = maskedAuth[:25] + "..."
		}
		log.Printf("[JWT-AUTH] Authorization: %s", maskedAuth)

		// 检查Authorization格式：Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("[JWT-AUTH] Authorization 格式错误, 路径: %s", path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			log.Printf("[JWT-AUTH] Token验证失败: %v, 路径: %s", err, path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		log.Printf("[JWT-AUTH] Token验证成功, 用户: %s, 路径: %s", claims.Username, path)

		// 将用户信息填入 RequestContext
		rc := GetRequestContext(c)
		if rc != nil {
			rc.Username = claims.Username
		}

		c.Next()
	}
}
