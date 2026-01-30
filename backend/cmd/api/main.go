package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/saas-starter-kit/backend/internal/api/handlers"
	"github.com/yourusername/saas-starter-kit/backend/internal/api/middleware"
	"github.com/yourusername/saas-starter-kit/backend/internal/config"
	"github.com/yourusername/saas-starter-kit/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := models.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed default plans
	if err := models.SeedPlans(db); err != nil {
		log.Fatalf("Failed to seed plans: %v", err)
	}

	// Create Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(middleware.CORS(cfg.FrontendURL))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	tenantHandler := handlers.NewTenantHandler(db, cfg)
	workspaceHandler := handlers.NewWorkspaceHandler(db, cfg)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			// Social OAuth
			auth.GET("/social/:provider/login", authHandler.InitiateOAuth)
			auth.POST("/social/callback", authHandler.HandleOAuthCallback)

			// Email/Password
			auth.POST("/register", authHandler.Register)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/login", authHandler.Login)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)

			// Protected
			auth.GET("/me", middleware.RequireAuth(cfg), authHandler.GetCurrentUser)
		}

		// Tenant routes (require auth)
		tenant := v1.Group("/tenant")
		tenant.Use(middleware.RequireAuth(cfg))
		{
			tenant.GET("", tenantHandler.GetCurrentTenant)
			tenant.GET("/plans", tenantHandler.ListPlans)
			tenant.POST("/select-plan", tenantHandler.SelectPlan)
			tenant.POST("/setup", tenantHandler.SetupOrganization)
			tenant.GET("/check-slug", tenantHandler.CheckSlug)
		}

		// Workspace routes (require auth + tenant)
		workspaces := v1.Group("/workspaces")
		workspaces.Use(middleware.RequireAuth(cfg))
		workspaces.Use(middleware.RequireTenant(db))
		{
			workspaces.GET("", workspaceHandler.List)
			workspaces.POST("", workspaceHandler.Create)
			workspaces.GET("/:id", workspaceHandler.Get)
			workspaces.DELETE("/:id", workspaceHandler.Delete)
			workspaces.GET("/:id/members", workspaceHandler.ListMembers)
			workspaces.POST("/:id/members", workspaceHandler.AddMember)
		}
	}

	// Start server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
