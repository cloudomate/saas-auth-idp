package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/authz-service/internal/auth"
	"github.com/yourusername/authz-service/internal/fga"
	"github.com/yourusername/authz-service/internal/handlers"
)

func main() {
	// Initialize Casdoor JWT validator
	casdoorEndpoint := getEnv("CASDOOR_ENDPOINT", "http://casdoor:8000")
	casdoorOrg := getEnv("CASDOOR_ORGANIZATION", "built-in")
	casdoorApp := getEnv("CASDOOR_APPLICATION", "app-built-in")

	jwtValidator, err := auth.NewCasdoorValidator(casdoorEndpoint, casdoorOrg, casdoorApp)
	if err != nil {
		log.Printf("Warning: Casdoor validator initialization failed: %v", err)
		log.Println("Running without JWT validation")
	} else {
		log.Printf("Casdoor validator initialized: %s", casdoorEndpoint)
	}

	// Initialize OpenFGA client
	fgaURL := getEnv("OPENFGA_URL", "http://openfga:8080")
	fgaStoreID := getStoreID()

	var fgaClient *fga.Client
	if fgaStoreID != "" {
		fgaClient, err = fga.NewClient(fgaURL, fgaStoreID)
		if err != nil {
			log.Printf("Warning: OpenFGA client initialization failed: %v", err)
		} else {
			log.Printf("OpenFGA client initialized with store: %s", fgaStoreID)
		}
	} else {
		log.Println("No OpenFGA store ID configured - permission checks disabled")
	}

	// Check for dev mode
	devMode := getEnv("DEV_MODE", "false") == "true"
	if devMode {
		log.Println("WARNING: Running in DEV_MODE - authentication bypassed!")
	}

	// Initialize handler
	gateHandler := handlers.NewGateHandler(jwtValidator, fgaClient, devMode)

	// Setup router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "authz"})
	})

	// ForwardAuth endpoint - called by Traefik for every request
	r.GET("/gate", gateHandler.Handle)
	r.POST("/gate", gateHandler.Handle)
	r.Any("/gate/*path", gateHandler.Handle)

	port := getEnv("PORT", "8002")
	log.Printf("AuthZ service starting on port %s", port)
	r.Run(":" + port)
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
