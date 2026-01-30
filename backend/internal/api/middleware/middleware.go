package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourusername/saas-starter-kit/backend/internal/config"
	"github.com/yourusername/saas-starter-kit/backend/internal/models"
	"gorm.io/gorm"
)

// CORS middleware for handling Cross-Origin requests
func CORS(frontendURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Allow the configured frontend URL and localhost for development
		allowedOrigins := []string{frontendURL, "http://localhost:5173", "http://localhost:3000"}
		allowed := false
		for _, o := range allowedOrigins {
			if origin == o {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Workspace-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Type          string `json:"type"` // "platform"
	EmailVerified bool   `json:"email_verified"`
	IsTenantAdmin bool   `json:"is_tenant_admin"`
	TenantID      string `json:"tenant_id,omitempty"`
	jwt.RegisteredClaims
}

// RequireAuth middleware validates JWT tokens
func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_token",
				"message": "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token_format",
				"message": "Authorization header must be Bearer {token}",
			})
			return
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return cfg.GetJWTSecret(), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_token",
				"message": "Token is invalid or expired",
			})
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_claims",
				"message": "Token claims are invalid",
			})
			return
		}

		// Set user info in context
		c.Set("user_id", claims.Sub)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("is_tenant_admin", claims.IsTenantAdmin)
		if claims.TenantID != "" {
			c.Set("tenant_id", claims.TenantID)
		}

		c.Next()
	}
}

// RequireTenant middleware ensures user has a tenant
func RequireTenant(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists || tenantID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "no_tenant",
				"message": "You must set up an organization first",
			})
			return
		}

		// Verify tenant exists
		tenantUUID, err := uuid.Parse(tenantID.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "invalid_tenant",
				"message": "Invalid tenant ID",
			})
			return
		}

		var tenant models.Tenant
		if err := db.First(&tenant, "id = ?", tenantUUID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "tenant_not_found",
				"message": "Tenant not found",
			})
			return
		}

		c.Set("tenant", &tenant)
		c.Next()
	}
}

// RequireTenantAdmin middleware ensures user is a tenant admin
func RequireTenantAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		isTenantAdmin, exists := c.Get("is_tenant_admin")
		if !exists || !isTenantAdmin.(bool) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "not_tenant_admin",
				"message": "Only tenant administrators can perform this action",
			})
			return
		}

		c.Next()
	}
}
