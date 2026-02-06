package proxy

import (
	"back/pkg/internal_token"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// InternalProxy 内部系统反向代理
// 接收通过外部 JWT 认证的请求，替换为内部 token 后转发到内部系统
type InternalProxy struct {
	tokenManager *internal_token.Manager
	targetURL    *url.URL
	proxy        *httputil.ReverseProxy
}

// NewInternalProxy 创建内部系统反向代理
func NewInternalProxy(tokenManager *internal_token.Manager, targetURL string) *InternalProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		panic(fmt.Sprintf("无效的内部系统 URL: %s", targetURL))
	}

	p := &InternalProxy{
		tokenManager: tokenManager,
		targetURL:    target,
	}

	p.proxy = &httputil.ReverseProxy{
		Director: p.director,
		ModifyResponse: func(resp *http.Response) error {
			// 移除可能泄露内部信息的 header
			resp.Header.Del("Server")
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, `{"code":502,"message":"internal system unavailable"}`)
		},
	}

	return p
}

// director 重写请求：替换目标 URL 和 Authorization header
func (p *InternalProxy) director(req *http.Request) {
	req.URL.Scheme = p.targetURL.Scheme
	req.URL.Host = p.targetURL.Host
	req.Host = p.targetURL.Host

	// 移除 /api/v1/internal 前缀，保留后面的路径
	// 例如: /api/v1/internal/users/123 → /users/123
	if path := req.URL.Path; len(path) > 0 {
		// path 在 Gin 中已经是 /*path 匹配的部分
		// 不需要额外处理，Gin 会把 /api/v1/internal/xxx 中的 /xxx 传给 *path
	}

	// token 替换在 Handler 中处理（director 无法返回 error）
}

// Handler 返回 Gin 处理函数
func (p *InternalProxy) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取内部 token
		token, err := p.tokenManager.GetToken()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    http.StatusServiceUnavailable,
				"message": "failed to authenticate with internal system",
			})
			c.Abort()
			return
		}

		// 从 Gin 路由参数获取实际路径
		targetPath := c.Param("path")
		if targetPath == "" {
			targetPath = "/"
		}

		// 修改请求路径和 header
		c.Request.URL.Path = targetPath
		c.Request.Header.Set("Authorization", "Bearer "+token)

		// 移除可能干扰内部系统的 header
		c.Request.Header.Del("Cookie")

		p.proxy.ServeHTTP(c.Writer, c.Request)
	}
}
