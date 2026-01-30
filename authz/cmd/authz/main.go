package main

import (
	"context"
	"log"
	"time"

	"saas-authz/internal/auth"
	"saas-authz/internal/authz"
	"saas-authz/internal/config"
	"saas-authz/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting AuthZ service on port %s", cfg.Port)
	log.Printf("OpenFGA URL: %s", cfg.OpenFGAURL)
	log.Printf("Dev mode: %v", cfg.DevMode)

	// Initialize JWT validator
	var jwtValidator *auth.JWTValidator
	if len(cfg.JWTSecret) > 0 {
		jwtValidator = auth.NewJWTValidator(cfg.JWTSecret)
		log.Printf("JWT validator initialized")
	} else {
		log.Printf("Warning: JWT secret not configured")
	}

	// Initialize API key validator
	var apiKeyValidator *auth.APIKeyValidator
	if cfg.DatabaseURL != "" && len(cfg.APIKeySecret) > 0 {
		var err error
		apiKeyValidator, err = auth.NewAPIKeyValidator(cfg.DatabaseURL, cfg.APIKeySecret)
		if err != nil {
			log.Printf("Warning: Failed to initialize API key validator: %v", err)
		} else {
			log.Printf("API key validator initialized")
		}
	} else {
		log.Printf("Warning: API key validation not configured")
	}

	// Initialize OpenFGA client
	openfgaClient := authz.NewClient(cfg.OpenFGAURL, cfg.OpenFGAStoreID, cfg.DevMode)
	if !cfg.DevMode && cfg.OpenFGAStoreID != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := openfgaClient.Initialize(ctx); err != nil {
			log.Printf("Warning: Failed to initialize OpenFGA client: %v", err)
		} else {
			log.Printf("OpenFGA client initialized")
		}
	}

	// Create handler
	gateHandler := handlers.NewGateHandler(jwtValidator, apiKeyValidator, openfgaClient, cfg.DevMode)

	// Setup Gin
	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ForwardAuth endpoint
	r.GET("/gate", gateHandler.Handle)
	r.POST("/gate", gateHandler.Handle)

	// Start server
	log.Printf("AuthZ service listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
