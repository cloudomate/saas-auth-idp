package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/authz-service/internal/auth"
	"github.com/yourusername/authz-service/internal/fga"
)

// GateHandler handles ForwardAuth requests from Traefik
type GateHandler struct {
	jwtValidator *auth.CasdoorValidator
	fgaClient    *fga.Client
	devMode      bool
}

// NewGateHandler creates a new gate handler
func NewGateHandler(jwtValidator *auth.CasdoorValidator, fgaClient *fga.Client, devMode bool) *GateHandler {
	return &GateHandler{
		jwtValidator: jwtValidator,
		fgaClient:    fgaClient,
		devMode:      devMode,
	}
}

// Handle processes ForwardAuth requests
// Traefik forwards the original request headers, and we validate the auth
// On success, we return 200 with headers that Traefik forwards to the backend
// On failure, we return 401/403
func (h *GateHandler) Handle(c *gin.Context) {
	// Get the original request path from Traefik headers
	originalURI := c.GetHeader("X-Forwarded-Uri")
	originalMethod := c.GetHeader("X-Forwarded-Method")

	log.Printf("Gate: %s %s", originalMethod, originalURI)

	// Check if this is a public endpoint
	if h.isPublicEndpoint(originalURI, originalMethod) {
		c.Status(http.StatusOK)
		return
	}

	// Dev mode - bypass authentication
	if h.devMode {
		h.setDevModeHeaders(c)
		c.Status(http.StatusOK)
		return
	}

	// Get the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "missing authorization header",
		})
		return
	}

	// Validate the JWT token
	if h.jwtValidator == nil {
		log.Println("Warning: JWT validator not configured")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "configuration_error",
			"message": "authentication service not configured",
		})
		return
	}

	userCtx, err := h.jwtValidator.ValidateToken(authHeader)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "invalid or expired token",
		})
		return
	}

	// Extract workspace from headers (set by frontend)
	workspaceID := c.GetHeader("X-Workspace-ID")

	// Check OpenFGA permissions if configured and workspace is specified
	if h.fgaClient != nil && workspaceID != "" && !userCtx.IsGlobalAdmin {
		allowed, err := h.checkPermission(c.Request.Context(), userCtx, workspaceID, originalMethod)
		if err != nil {
			log.Printf("Permission check error: %v", err)
			// Allow on error (fail open) - you might want to fail closed in production
		} else if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
			return
		}
	}

	// Set response headers for downstream services
	h.setUserHeaders(c, userCtx, workspaceID)

	c.Status(http.StatusOK)
}

// isPublicEndpoint checks if the endpoint doesn't require authentication
func (h *GateHandler) isPublicEndpoint(uri, method string) bool {
	// Health check
	if uri == "/health" || uri == "/api/health" {
		return true
	}

	// Only specific auth endpoints are public
	publicAuthEndpoints := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/callback",
		"/api/v1/auth/config",
		"/api/v1/auth/logout",
		"/api/v1/auth/social/",
	}

	for _, endpoint := range publicAuthEndpoints {
		if strings.HasPrefix(uri, endpoint) {
			return true
		}
	}

	return false
}

// checkPermission checks if the user has permission for the requested action
func (h *GateHandler) checkPermission(ctx context.Context, userCtx *auth.UserContext, workspaceID, method string) (bool, error) {
	user := "user:" + userCtx.UserID
	object := "container:" + workspaceID

	// Map HTTP methods to OpenFGA relations
	var relation string
	switch method {
	case "GET", "HEAD", "OPTIONS":
		relation = "can_read"
	case "POST", "PUT", "PATCH":
		relation = "can_write"
	case "DELETE":
		relation = "can_manage"
	default:
		relation = "can_read"
	}

	return h.fgaClient.Check(ctx, user, relation, object)
}

// setUserHeaders sets headers for downstream services
func (h *GateHandler) setUserHeaders(c *gin.Context, userCtx *auth.UserContext, workspaceID string) {
	c.Header("X-User-ID", userCtx.UserID)
	c.Header("X-User-Name", userCtx.Name)
	c.Header("X-User-Email", userCtx.Email)
	c.Header("X-Tenant-ID", userCtx.Organization)
	c.Header("X-Is-Admin", boolToString(userCtx.IsAdmin))
	c.Header("X-Is-Platform-Admin", boolToString(userCtx.IsGlobalAdmin))

	if workspaceID != "" {
		c.Header("X-Workspace-ID", workspaceID)
	}

	// Pass through the original Authorization header for downstream services
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		c.Header("Authorization", authHeader)
	}
}

// setDevModeHeaders sets default headers for development mode
func (h *GateHandler) setDevModeHeaders(c *gin.Context) {
	c.Header("X-User-ID", "dev-user")
	c.Header("X-User-Name", "Development User")
	c.Header("X-User-Email", "dev@example.com")
	c.Header("X-Tenant-ID", "built-in")
	c.Header("X-Is-Admin", "true")
	c.Header("X-Is-Platform-Admin", "true")
	c.Header("X-Workspace-ID", "workspace-default")
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
