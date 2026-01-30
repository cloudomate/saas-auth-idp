package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/sample-api/internal/casdoor"
	"github.com/yourusername/sample-api/internal/middleware"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	casdoorClient *casdoor.Client
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(client *casdoor.Client) *AuthHandler {
	return &AuthHandler{
		casdoorClient: client,
	}
}

// ConfigResponse represents the auth configuration for the frontend
type ConfigResponse struct {
	AuthEndpoint string `json:"auth_endpoint"`
	ClientID     string `json:"client_id"`
	Organization string `json:"organization"`
	Application  string `json:"application"`
	RedirectURI  string `json:"redirect_uri"`
	AuthMode     string `json:"auth_mode"` // "headless" or "redirect"
}

// GetConfig returns the auth configuration for the frontend
func (h *AuthHandler) GetConfig(c *gin.Context) {
	endpoint := getEnv("AUTH_ENDPOINT", "http://localhost:4455")

	config := ConfigResponse{
		AuthEndpoint: endpoint,
		ClientID:     os.Getenv("CASDOOR_CLIENT_ID"),
		Organization: getEnv("CASDOOR_ORGANIZATION", "built-in"),
		Application:  getEnv("CASDOOR_APPLICATION", "app-built-in"),
		RedirectURI:  getEnv("APP_URL", "http://localhost:3000") + "/callback",
		AuthMode:     "headless",
	}

	c.JSON(http.StatusOK, config)
}

// LoginRequest represents the headless login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response with tokens
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// Login handles headless password authentication via IDP API
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "email and password are required",
		})
		return
	}

	// Call IDP's login API with type="token" to get JWT directly
	idpEndpoint := getEnv("CASDOOR_ENDPOINT", "http://casdoor:8000")
	org := getEnv("CASDOOR_ORGANIZATION", "saas-platform")
	app := getEnv("CASDOOR_APPLICATION", "saas-app")

	// Login with type="token" to get JWT directly
	loginPayload := map[string]interface{}{
		"application":  app,
		"organization": org,
		"username":     req.Email,
		"password":     req.Password,
		"autoSignin":   true,
		"type":         "token",
	}

	jsonData, err := json.Marshal(loginPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to prepare request",
		})
		return
	}

	resp, err := http.Post(idpEndpoint+"/api/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service_unavailable",
			"message": "failed to connect to identity provider",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var loginResp map[string]interface{}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to parse response",
		})
		return
	}

	// Check if login was successful
	status, ok := loginResp["status"].(string)
	if !ok || status != "ok" {
		msg := "authentication failed"
		if m, ok := loginResp["msg"].(string); ok && m != "" {
			msg = m
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": msg,
		})
		return
	}

	// Get the JWT token from data field
	accessToken, ok := loginResp["data"].(string)
	if !ok || accessToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "no access token in response",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   86400, // 24 hours (configurable in IDP)
	})
}

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name"`
	Phone       string `json:"phone"`
}

// Register handles headless user registration via IDP API
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "email and password are required",
		})
		return
	}

	idpEndpoint := getEnv("CASDOOR_ENDPOINT", "http://casdoor:8000")
	org := getEnv("CASDOOR_ORGANIZATION", "built-in")
	app := getEnv("CASDOOR_APPLICATION", "app-built-in")

	// Prepare user data for IDP signup - use email as username
	signupPayload := map[string]interface{}{
		"application":  app,
		"organization": org,
		"username":     req.Email, // Use email as username
		"password":     req.Password,
		"name":         req.Email, // Use email as name
		"email":        req.Email,
		"displayName":  req.DisplayName,
		"phone":        req.Phone,
		"type":         "signup",
	}

	jsonData, err := json.Marshal(signupPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to prepare request",
		})
		return
	}

	resp, err := http.Post(idpEndpoint+"/api/signup", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service_unavailable",
			"message": "failed to connect to identity provider",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var idpResp map[string]interface{}
	if err := json.Unmarshal(body, &idpResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to parse response",
		})
		return
	}

	// Check if signup was successful
	status, ok := idpResp["status"].(string)
	if !ok || status != "ok" {
		msg := "registration failed"
		if m, ok := idpResp["msg"].(string); ok && m != "" {
			msg = m
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "registration_failed",
			"message": msg,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"status":  "ok",
	})
}

// GetSocialLoginURL returns the OAuth URL for a social provider
func (h *AuthHandler) GetSocialLoginURL(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "provider is required",
		})
		return
	}

	endpoint := getEnv("AUTH_ENDPOINT", "http://localhost:4455")
	clientID := os.Getenv("CASDOOR_CLIENT_ID")
	org := getEnv("CASDOOR_ORGANIZATION", "built-in")
	app := getEnv("CASDOOR_APPLICATION", "app-built-in")
	redirectURI := getEnv("APP_URL", "http://localhost:3000") + "/callback"

	// Construct OAuth URL with provider hint
	authURL := fmt.Sprintf(
		"%s/login/oauth/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=openid+profile+email&state=%s&provider=%s",
		endpoint, clientID, redirectURI, org+"/"+app, provider,
	)

	c.JSON(http.StatusOK, gin.H{
		"url":      authURL,
		"provider": provider,
	})
}

// CallbackRequest represents the OAuth callback request
type CallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

// Callback handles the OAuth callback - exchanges code for token
func (h *AuthHandler) Callback(c *gin.Context) {
	var req CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "code is required",
		})
		return
	}

	idpEndpoint := getEnv("CASDOOR_ENDPOINT", "http://casdoor:8000")
	clientID := os.Getenv("CASDOOR_CLIENT_ID")
	clientSecret := os.Getenv("CASDOOR_CLIENT_SECRET")
	redirectURI := getEnv("APP_URL", "http://localhost:3000") + "/callback"

	// Exchange code for token
	tokenPayload := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          req.Code,
		"redirect_uri":  redirectURI,
	}

	jsonData, _ := json.Marshal(tokenPayload)
	resp, err := http.Post(
		idpEndpoint+"/api/login/oauth/access_token",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service_unavailable",
			"message": "failed to exchange code for token",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to parse token response",
		})
		return
	}

	// Check for error
	if errMsg, ok := tokenResp["error"].(string); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   errMsg,
			"message": tokenResp["error_description"],
		})
		return
	}

	c.JSON(http.StatusOK, tokenResp)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out",
		"status":  "ok",
	})
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword handles password change via IDP API
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "old_password and new_password are required",
		})
		return
	}

	idpEndpoint := getEnv("CASDOOR_ENDPOINT", "http://casdoor:8000")
	org := getEnv("CASDOOR_ORGANIZATION", "saas-platform")

	// Get the original Authorization header to pass to Casdoor
	authHeader := c.GetHeader("Authorization")

	// Try to get user info from Casdoor claims first (direct mode)
	var userID string
	var userOrg string
	if claims, exists := c.Get(middleware.CasdoorClaimsKey); exists {
		idpClaims := claims.(*casdoor.CasdoorClaims)
		userID = idpClaims.Name
		userOrg = idpClaims.Owner
	} else {
		// Try to get from middleware.UserContext (gateway mode)
		userCtx := middleware.GetUserContext(c)
		if userCtx != nil && userCtx.UserID != "" {
			userID = userCtx.UserID
			if userCtx.TenantID != "" {
				userOrg = userCtx.TenantID
			}
		}
	}

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "user not authenticated",
		})
		return
	}

	// Use org from claims if available, otherwise use default
	if userOrg != "" {
		org = userOrg
	}

	// Log for debugging
	fmt.Printf("ChangePassword: org=%s, userName=%s\n", org, userID)

	// Casdoor's set-password API uses form data, not JSON
	formData := fmt.Sprintf("userOwner=%s&userName=%s&oldPassword=%s&newPassword=%s",
		org, userID, req.OldPassword, req.NewPassword)

	httpReq, err := http.NewRequest("POST", idpEndpoint+"/api/set-password", strings.NewReader(formData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to prepare request",
		})
		return
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Pass the user's token for authentication
	if authHeader != "" {
		httpReq.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service_unavailable",
			"message": "failed to connect to identity provider",
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var idpResp map[string]interface{}
	if err := json.Unmarshal(body, &idpResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to parse response",
		})
		return
	}

	// Check if password change was successful
	status, ok := idpResp["status"].(string)
	if !ok || status != "ok" {
		msg := "password change failed"
		if m, ok := idpResp["msg"].(string); ok && m != "" {
			msg = m
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "password_change_failed",
			"message": msg,
		})
		return
	}

	// Clear passwordChangeRequired property in Casdoor
	// Use client credentials for this admin operation
	clientID := os.Getenv("CASDOOR_CLIENT_ID")
	clientSecret := os.Getenv("CASDOOR_CLIENT_SECRET")
	clearPasswordChangeRequired(idpEndpoint, org, userID, clientID, clientSecret)

	c.JSON(http.StatusOK, gin.H{
		"message": "password changed successfully",
		"status":  "ok",
	})
}

// clearPasswordChangeRequired updates the user's properties to remove the password change requirement
func clearPasswordChangeRequired(idpEndpoint, org, userName, clientID, clientSecret string) {
	// Get the current user first using client credentials
	getUserURL := fmt.Sprintf("%s/api/get-user?id=%s/%s", idpEndpoint, org, userName)

	req, err := http.NewRequest("GET", getUserURL, nil)
	if err != nil {
		fmt.Printf("clearPasswordChangeRequired: failed to create get request: %v\n", err)
		return
	}
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("clearPasswordChangeRequired: failed to get user: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Parse the response - user data is in "data" field
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		fmt.Printf("clearPasswordChangeRequired: failed to parse response: %v\n", err)
		return
	}

	user, ok := apiResponse["data"].(map[string]interface{})
	if !ok {
		fmt.Printf("clearPasswordChangeRequired: no user data in response\n")
		return
	}

	// Update the properties to remove passwordChangeRequired
	properties, ok := user["properties"].(map[string]interface{})
	if !ok {
		properties = make(map[string]interface{})
	}
	delete(properties, "passwordChangeRequired")
	user["properties"] = properties

	// Update the user
	updatePayload, _ := json.Marshal(user)
	updateURL := fmt.Sprintf("%s/api/update-user?id=%s/%s", idpEndpoint, org, userName)
	updateReq, _ := http.NewRequest("POST", updateURL, bytes.NewBuffer(updatePayload))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.SetBasicAuth(clientID, clientSecret)

	updateResp, err := client.Do(updateReq)
	if err != nil {
		fmt.Printf("clearPasswordChangeRequired: failed to update user: %v\n", err)
		return
	}
	defer updateResp.Body.Close()

	updateBody, _ := io.ReadAll(updateResp.Body)
	fmt.Printf("clearPasswordChangeRequired: update response: %s\n", string(updateBody))
}

// UserResponse represents the current user info
type UserResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	Email         string `json:"email"`
	Avatar        string `json:"avatar"`
	Organization  string `json:"organization"`
	IsAdmin       bool   `json:"is_admin"`
	IsGlobalAdmin bool   `json:"is_global_admin"`
}

// GetMe returns the current user's info from the validated token
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Try to get claims from IDP middleware first
	claims, exists := c.Get(middleware.CasdoorClaimsKey)
	if exists {
		idpClaims := claims.(*casdoor.CasdoorClaims)

		user := UserResponse{
			ID:            idpClaims.Name,
			Name:          idpClaims.Name,
			DisplayName:   idpClaims.DisplayName,
			Email:         idpClaims.Email,
			Avatar:        idpClaims.Avatar,
			Organization:  idpClaims.Owner,
			IsAdmin:       idpClaims.IsAdmin,
			IsGlobalAdmin: idpClaims.IsGlobalAdmin,
		}

		c.JSON(http.StatusOK, user)
		return
	}

	// Fall back to gateway headers
	userCtx := middleware.GetUserContext(c)
	if userCtx != nil && userCtx.UserID != "" && userCtx.UserID != "anonymous" {
		c.JSON(http.StatusOK, UserResponse{
			ID:            userCtx.UserID,
			Name:          userCtx.UserID,
			Organization:  userCtx.TenantID,
			IsGlobalAdmin: userCtx.IsPlatformAdmin,
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{
		"error":   "unauthorized",
		"message": "no valid token",
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
