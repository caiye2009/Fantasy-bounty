package middleware

import (
	"back/internal/audit"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Audit returns a gin middleware that logs completed requests to the audit service.
func Audit(auditService audit.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Let the handler chain execute first
		c.Next()

		rc := GetRequestContext(c)
		if rc == nil {
			return
		}

		action := rc.Action
		if action == "" {
			action = autoAction(c.Request.Method, c.FullPath())
		}

		var detailJSON string
		if rc.Detail != nil {
			b, err := json.Marshal(rc.Detail)
			if err == nil {
				detailJSON = string(b)
			} else {
				detailJSON = "null"
			}
		} else {
			detailJSON = "null"
		}

		entry := &audit.AuditLog{
			RequestID:  rc.RequestID,
			Username:   rc.Username,
			Action:     action,
			Resource:   rc.Resource,
			ResourceID: rc.ResourceID,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			ClientIP:   rc.ClientIP,
			UserAgent:  rc.UserAgent,
			Duration:   time.Since(rc.StartTime).Milliseconds(),
			Detail:     detailJSON,
		}

		auditService.Log(entry)
	}
}

// autoAction generates a default action string from the HTTP method and route path.
// Example: GET /api/v1/suppliers -> supplier.list
func autoAction(method, path string) string {
	if path == "" {
		return fmt.Sprintf("%s:%s", method, path)
	}

	// Strip /api/v1/ prefix
	trimmed := strings.TrimPrefix(path, "/api/v1/")

	// Split into segments
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return fmt.Sprintf("%s:%s", method, path)
	}

	// Take the first segment as resource name, singularise naively
	resource := parts[0]
	if strings.HasSuffix(resource, "ies") {
		resource = strings.TrimSuffix(resource, "ies") + "y"
	} else if strings.HasSuffix(resource, "s") {
		resource = strings.TrimSuffix(resource, "s")
	}

	// Determine verb from method + path shape
	hasID := len(parts) >= 2 && strings.HasPrefix(parts[1], ":")
	subResource := ""
	if len(parts) >= 2 && !strings.HasPrefix(parts[1], ":") {
		subResource = parts[1]
	}
	if hasID && len(parts) >= 3 {
		subResource = parts[2]
	}

	var verb string
	switch method {
	case "GET":
		if hasID || subResource == "my" {
			verb = "get"
		} else {
			verb = "list"
		}
	case "POST":
		verb = "create"
	case "PUT", "PATCH":
		verb = "update"
	case "DELETE":
		verb = "delete"
	default:
		verb = strings.ToLower(method)
	}

	if subResource != "" && subResource != "my" {
		return fmt.Sprintf("%s.%s_%s", resource, verb, subResource)
	}
	return fmt.Sprintf("%s.%s", resource, verb)
}
