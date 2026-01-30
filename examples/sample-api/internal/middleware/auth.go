package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/yourusername/sample-api/internal/store"
)

const (
	UserContextKey = "user_context"
)

// ExtractAuthHeaders extracts authentication headers set by the AuthZ service
// These headers are set by the ForwardAuth middleware after validating the JWT
func ExtractAuthHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtx := store.UserContext{
			UserID:          c.GetHeader("X-User-ID"),
			TenantID:        c.GetHeader("X-Tenant-ID"),
			WorkspaceID:     c.GetHeader("X-Workspace-ID"),
			IsPlatformAdmin: c.GetHeader("X-Is-Platform-Admin") == "true",
		}

		// For development, allow passing user context via query params
		if userCtx.UserID == "" {
			userCtx.UserID = c.Query("user_id")
			userCtx.TenantID = c.Query("tenant_id")
			userCtx.WorkspaceID = c.Query("workspace_id")
		}

		// Default values for demo
		if userCtx.UserID == "" {
			userCtx.UserID = "user-1"
		}
		if userCtx.WorkspaceID == "" {
			userCtx.WorkspaceID = "workspace-1"
		}

		c.Set(UserContextKey, &userCtx)
		c.Next()
	}
}

// GetUserContext retrieves the user context from gin context
func GetUserContext(c *gin.Context) *store.UserContext {
	if ctx, exists := c.Get(UserContextKey); exists {
		return ctx.(*store.UserContext)
	}
	return nil
}
