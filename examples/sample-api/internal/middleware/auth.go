package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/sample-api/internal/casdoor"
	"github.com/yourusername/sample-api/internal/store"
)

const (
	UserContextKey    = "user_context"
	CasdoorClaimsKey  = "casdoor_claims"
)

// CasdoorAuth validates Casdoor JWT tokens
func CasdoorAuth(client *casdoor.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Check for token in query param (for WebSocket/SSE connections)
			authHeader = c.Query("token")
		}

		var userCtx store.UserContext

		if authHeader != "" {
			// Validate the token
			claims, err := client.ValidateToken(authHeader)
			if err != nil {
				log.Printf("Token validation failed: %v", err)
				c.AbortWithStatusJSON(401, gin.H{
					"error":   "unauthorized",
					"message": "invalid or expired token",
				})
				return
			}

			// Store claims in context
			c.Set(CasdoorClaimsKey, claims)

			// Build user context from Casdoor claims
			userCtx = store.UserContext{
				UserID:          claims.Name, // Casdoor uses "name" as the user identifier
				TenantID:        claims.Owner, // Organization is the tenant
				IsPlatformAdmin: claims.IsGlobalAdmin,
			}

			// Get workspace from header or query param
			userCtx.WorkspaceID = c.GetHeader("X-Workspace-ID")
			if userCtx.WorkspaceID == "" {
				userCtx.WorkspaceID = c.Query("workspace_id")
			}
		} else {
			// No auth header - use development defaults if DEV_MODE is enabled
			// In production, you would reject the request here
			userCtx = store.UserContext{
				UserID:      "anonymous",
				WorkspaceID: c.Query("workspace_id"),
			}
		}

		c.Set(UserContextKey, &userCtx)
		c.Next()
	}
}

// ExtractAuthHeaders extracts authentication headers set by the AuthZ service
// These headers are set by the ForwardAuth middleware after validating the JWT
// Used when running behind the authz service
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

// RequireAuth middleware ensures the user is authenticated
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtx := GetUserContext(c)
		if userCtx == nil || userCtx.UserID == "" || userCtx.UserID == "anonymous" {
			c.AbortWithStatusJSON(401, gin.H{
				"error":   "unauthorized",
				"message": "authentication required",
			})
			return
		}
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

// RequirePlatformAdmin middleware ensures the user is a platform admin
func RequirePlatformAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userCtx := GetUserContext(c)
		if userCtx == nil || !userCtx.IsPlatformAdmin {
			c.AbortWithStatusJSON(403, gin.H{
				"error":   "forbidden",
				"message": "platform admin access required",
			})
			return
		}
		c.Next()
	}
}
