package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/yourusername/sample-api/internal/authz"
	"github.com/yourusername/sample-api/internal/casdoor"
	"github.com/yourusername/sample-api/internal/handlers"
	"github.com/yourusername/sample-api/internal/middleware"
	"github.com/yourusername/sample-api/internal/store"
)

func main() {
	// Initialize OpenFGA client
	fgaURL := getEnv("OPENFGA_URL", "http://localhost:8081")
	fgaStoreID := getStoreID()

	var fgaClient *authz.OpenFGAClient
	var err error

	if fgaStoreID != "" {
		fgaClient, err = authz.NewOpenFGAClient(fgaURL, fgaStoreID)
		if err != nil {
			log.Printf("Warning: OpenFGA client initialization failed: %v", err)
			log.Println("Running without OpenFGA - using mock authorization")
		} else {
			log.Printf("OpenFGA client initialized with store: %s", fgaStoreID)
		}
	} else {
		log.Println("No OpenFGA store ID configured - using mock authorization")
	}

	// Check auth mode
	authMode := getEnv("AUTH_MODE", "gateway")
	log.Printf("Auth mode: %s", authMode)

	// Initialize Casdoor client (only needed for direct mode)
	var casdoorClient *casdoor.Client
	if authMode == "direct" {
		casdoorClient, err = casdoor.NewClientFromEnv()
		if err != nil {
			log.Printf("Warning: Casdoor client initialization failed: %v", err)
			log.Println("Falling back to header-based auth")
		} else {
			log.Printf("Casdoor client initialized: %s", getEnv("CASDOOR_ENDPOINT", "http://localhost:8100"))
		}
	} else {
		log.Println("Gateway mode: trusting headers from Traefik/AuthZ service")
	}

	// Initialize in-memory store (replace with real DB in production)
	dataStore := store.NewMemoryStore()

	// Seed sample data (Casdoor manages users, we just need tenant/workspace data)
	seedSampleData(dataStore)
	seedData(dataStore)

	// Initialize handlers
	docHandler := handlers.NewDocumentHandler(dataStore, fgaClient)
	projectHandler := handlers.NewProjectHandler(dataStore, fgaClient)
	adminHandler := handlers.NewAdminHandler(dataStore)
	authHandler := handlers.NewAuthHandler(casdoorClient)

	// Setup router
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:5173", "http://localhost:4455"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-User-ID", "X-Tenant-ID", "X-Workspace-ID"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth routes (public - headless mode)
	authRoutes := r.Group("/api/v1/auth")
	{
		authRoutes.GET("/config", authHandler.GetConfig)
		authRoutes.POST("/login", authHandler.Login)           // Headless login
		authRoutes.POST("/register", authHandler.Register)     // Headless registration
		authRoutes.POST("/callback", authHandler.Callback)     // OAuth code exchange
		authRoutes.POST("/logout", authHandler.Logout)
		authRoutes.GET("/social/:provider", authHandler.GetSocialLoginURL) // Get OAuth URL
		// Protected auth routes (require authentication)
		if authMode == "direct" && casdoorClient != nil {
			authRoutes.GET("/me", middleware.CasdoorAuth(casdoorClient), authHandler.GetMe)
			authRoutes.POST("/change-password", middleware.CasdoorAuth(casdoorClient), authHandler.ChangePassword)
		} else {
			authRoutes.GET("/me", middleware.ExtractAuthHeaders(), authHandler.GetMe)
			authRoutes.POST("/change-password", middleware.ExtractAuthHeaders(), authHandler.ChangePassword)
		}
	}

	// API routes - use appropriate auth middleware based on mode
	api := r.Group("/api/v1")
	if authMode == "direct" && casdoorClient != nil {
		// Direct mode: validate Casdoor JWT in this service
		api.Use(middleware.CasdoorAuth(casdoorClient))
		log.Println("API using direct Casdoor JWT validation")
	} else {
		// Gateway mode: trust headers from Traefik/AuthZ
		api.Use(middleware.ExtractAuthHeaders())
		log.Println("API using gateway headers (X-User-ID, X-Tenant-ID, etc.)")
	}
	{
		// Document routes (ReBAC example)
		docs := api.Group("/documents")
		{
			docs.GET("", docHandler.List)
			docs.POST("", docHandler.Create)
			docs.GET("/:id", docHandler.Get)
			docs.PUT("/:id", docHandler.Update)
			docs.DELETE("/:id", docHandler.Delete)
			docs.POST("/:id/share", docHandler.Share)
			docs.GET("/:id/permissions", docHandler.GetPermissions)
		}

		// Project routes (ABAC example)
		projects := api.Group("/projects")
		{
			projects.GET("", projectHandler.List)
			projects.POST("", projectHandler.Create)
			projects.GET("/:id", projectHandler.Get)
			projects.PUT("/:id", projectHandler.Update)
			projects.DELETE("/:id", projectHandler.Delete)
			projects.POST("/:id/deploy", projectHandler.Deploy)
		}

		// Permission check endpoint
		api.POST("/check-permission", func(c *gin.Context) {
			var req struct {
				User     string `json:"user"`
				Relation string `json:"relation"`
				Object   string `json:"object"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}

			if fgaClient == nil {
				c.JSON(http.StatusOK, gin.H{"allowed": true, "mock": true})
				return
			}

			allowed, err := fgaClient.Check(req.User, req.Relation, req.Object)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"allowed": allowed})
		})

		// Admin routes (require platform admin)
		admin := api.Group("/admin")
		admin.Use(middleware.RequirePlatformAdmin())
		{
			// Platform stats
			admin.GET("/stats", adminHandler.GetStats)

			// User management
			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/users/:id", adminHandler.GetUser)
			admin.PUT("/users/:id/admin", adminHandler.UpdateUserAdmin)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)

			// Tenant management
			admin.GET("/tenants", adminHandler.ListTenants)
			admin.GET("/tenants/:id", adminHandler.GetTenant)
			admin.DELETE("/tenants/:id", adminHandler.DeleteTenant)

			// Workspace management
			admin.GET("/workspaces", adminHandler.ListWorkspaces)
			admin.DELETE("/workspaces/:id", adminHandler.DeleteWorkspace)

			// View all resources
			admin.GET("/documents", adminHandler.ListAllDocuments)
			admin.GET("/projects", adminHandler.ListAllProjects)
		}
	}

	port := getEnv("PORT", "8001")
	log.Printf("Sample API server starting on port %s", port)
	r.Run(":" + port)
}

func seedSampleData(s *store.MemoryStore) {
	// Create sample tenants (organizations in Casdoor map to tenants here)
	tenants := []*store.Tenant{
		{
			ID:      "built-in",
			Name:    "Platform (Built-in)",
			Slug:    "built-in",
			Plan:    "enterprise",
			OwnerID: "admin",
		},
	}

	for _, tenant := range tenants {
		if err := s.CreateTenant(tenant); err != nil {
			log.Printf("Warning: Could not create tenant %s: %v", tenant.Name, err)
		}
	}

	// Create sample workspaces
	workspaces := []*store.Workspace{
		{
			ID:       "workspace-default",
			Name:     "Default Workspace",
			TenantID: "built-in",
		},
	}

	for _, workspace := range workspaces {
		if err := s.CreateWorkspace(workspace); err != nil {
			log.Printf("Warning: Could not create workspace %s: %v", workspace.Name, err)
		}
	}

	log.Printf("Seeded %d tenants, %d workspaces", len(tenants), len(workspaces))
}

func seedData(s *store.MemoryStore) {
	// Seed some sample documents
	s.CreateDocument(&store.Document{
		ID:          "doc-1",
		Title:       "Public Roadmap",
		Content:     "Q1: Feature A, Q2: Feature B...",
		WorkspaceID: "workspace-1",
		OwnerID:     "user-1",
		Visibility:  "workspace",
		Status:      "published",
	})

	s.CreateDocument(&store.Document{
		ID:          "doc-2",
		Title:       "Architecture Design",
		Content:     "System architecture overview...",
		WorkspaceID: "workspace-1",
		OwnerID:     "user-1",
		Visibility:  "private",
		Status:      "draft",
	})

	s.CreateDocument(&store.Document{
		ID:          "doc-3",
		Title:       "API Documentation",
		Content:     "REST API endpoints...",
		WorkspaceID: "workspace-1",
		OwnerID:     "user-2",
		Visibility:  "public",
		Status:      "published",
	})

	// Seed some sample projects
	s.CreateProject(&store.Project{
		ID:          "proj-1",
		Name:        "Web App",
		Description: "Main web application",
		WorkspaceID: "workspace-1",
		OwnerID:     "user-1",
		Environment: "production",
		Status:      "active",
		Tags:        []string{"frontend", "react"},
	})

	s.CreateProject(&store.Project{
		ID:          "proj-2",
		Name:        "API Service",
		Description: "Backend API service",
		WorkspaceID: "workspace-1",
		OwnerID:     "user-1",
		Environment: "staging",
		Status:      "active",
		Tags:        []string{"backend", "go"},
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getStoreID() string {
	// First check for direct environment variable
	if storeID := os.Getenv("OPENFGA_STORE_ID"); storeID != "" {
		return storeID
	}

	// Check for file-based store ID (used in Docker)
	storeIDFile := os.Getenv("OPENFGA_STORE_ID_FILE")
	if storeIDFile == "" {
		return ""
	}

	// Wait for file to appear (setup might still be running)
	log.Printf("Waiting for OpenFGA store ID file: %s", storeIDFile)
	for i := 0; i < 30; i++ {
		if data, err := os.ReadFile(storeIDFile); err == nil {
			storeID := strings.TrimSpace(string(data))
			if storeID != "" {
				log.Printf("Read store ID from file: %s", storeID)
				return storeID
			}
		}
		time.Sleep(1 * time.Second)
	}

	log.Printf("Warning: Could not read store ID from file after 30 seconds")
	return ""
}
