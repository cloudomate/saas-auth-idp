package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"saas-authz/internal/auth"
	"saas-authz/internal/authz"

	"github.com/gin-gonic/gin"
)

// GateHandler handles Traefik ForwardAuth requests
type GateHandler struct {
	jwt     *auth.JWTValidator
	apiKey  *auth.APIKeyValidator
	authz   *authz.Client
	devMode bool
}

// NewGateHandler creates a new gate handler
func NewGateHandler(jwt *auth.JWTValidator, apiKey *auth.APIKeyValidator, authzClient *authz.Client, devMode bool) *GateHandler {
	return &GateHandler{
		jwt:     jwt,
		apiKey:  apiKey,
		authz:   authzClient,
		devMode: devMode,
	}
}

// Handle processes ForwardAuth requests from Traefik
func (h *GateHandler) Handle(c *gin.Context) {
	originalMethod := c.GetHeader("X-Forwarded-Method")
	originalURI := c.GetHeader("X-Forwarded-Uri")
	authHeader := c.GetHeader("Authorization")

	log.Printf("[gate] Request: method=%s uri=%s auth=%v", originalMethod, originalURI, authHeader != "")

	// Dev mode bypass
	if h.devMode && authHeader == "" {
		log.Printf("[gate] Dev mode: allowing unauthenticated request")
		c.Header("X-User-ID", "00000000-0000-0000-0000-000000000001")
		c.Header("X-User-Email", "dev@localhost")
		c.Header("X-Tenant-ID", "00000000-0000-0000-0000-000000000001")
		c.Header("X-Is-Platform-Admin", "true")
		if wsID := c.GetHeader("X-Workspace-ID"); wsID != "" {
			c.Header("X-Workspace-ID", wsID)
		}
		c.Status(http.StatusOK)
		return
	}

	// Check for public routes
	if isPublicRoute(originalURI) {
		log.Printf("[gate] Public route: %s", originalURI)
		if authHeader != "" {
			identity, _ := h.authenticate(authHeader)
			if identity != nil {
				h.setResponseHeaders(c, identity)
			}
		}
		c.Status(http.StatusOK)
		return
	}

	// Authenticate
	if authHeader == "" {
		log.Printf("[gate] No authorization header")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	identity, err := h.authenticate(authHeader)
	if err != nil {
		log.Printf("[gate] Authentication failed: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Get workspace from header if not in token
	if identity.WorkspaceID == "" {
		identity.WorkspaceID = c.GetHeader("X-Workspace-ID")
	}

	// Authorize via OpenFGA (if workspace scoped)
	if identity.WorkspaceID != "" && !identity.IsPlatformAdmin {
		permission := methodToPermission(originalMethod)
		ctx := context.Background()

		allowed, err := h.authz.Check(ctx, identity.UserID, identity.WorkspaceID, permission, originalURI)
		if err != nil {
			log.Printf("[gate] Authorization check failed: %v", err)
		} else if !allowed {
			log.Printf("[gate] Authorization denied: user=%s workspace=%s permission=%s", identity.UserID, identity.WorkspaceID, permission)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}

	log.Printf("[gate] Authorized: user=%s email=%s tenant=%s workspace=%s admin=%v",
		identity.UserID, identity.Email, identity.TenantID, identity.WorkspaceID, identity.IsPlatformAdmin)

	h.setResponseHeaders(c, identity)
	c.Status(http.StatusOK)
}

func (h *GateHandler) authenticate(authHeader string) (*auth.Identity, error) {
	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimPrefix(token, "bearer ")

	// API key authentication (sk- prefix)
	if strings.HasPrefix(token, "sk-") {
		if h.apiKey == nil {
			return nil, fmt.Errorf("API key validation not configured")
		}
		return h.apiKey.Validate(token)
	}

	// JWT authentication
	if h.jwt == nil {
		return nil, fmt.Errorf("JWT validation not configured")
	}
	return h.jwt.Validate(token)
}

func (h *GateHandler) setResponseHeaders(c *gin.Context, id *auth.Identity) {
	c.Header("X-User-ID", id.UserID)
	c.Header("X-User-Email", id.Email)
	c.Header("X-Tenant-ID", id.TenantID)
	c.Header("X-Workspace-ID", id.WorkspaceID)
	c.Header("X-Role", id.Role)
	c.Header("X-Is-Platform-Admin", fmt.Sprintf("%v", id.IsPlatformAdmin))
	if id.KeyID != "" {
		c.Header("X-API-Key-ID", id.KeyID)
	}
}

func isPublicRoute(uri string) bool {
	publicPrefixes := []string{
		"/api/v1/health",
		"/api/v1/auth/",
		"/api/v1/tenant/plans",
		"/health",
	}
	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(uri, prefix) {
			return true
		}
	}
	return false
}

func methodToPermission(method string) string {
	switch method {
	case "GET", "HEAD", "OPTIONS":
		return "can_read"
	case "POST", "PUT", "PATCH":
		return "can_write"
	case "DELETE":
		return "can_manage"
	default:
		return "can_read"
	}
}
