package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/saas-starter-kit/backend/internal/config"
	"github.com/yourusername/saas-starter-kit/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

// ============================================================================
// OAuth Login
// ============================================================================

// InitiateOAuth starts the OAuth flow
// GET /api/v1/auth/social/:provider/login
func (h *AuthHandler) InitiateOAuth(c *gin.Context) {
	provider := c.Param("provider")
	_ = c.Query("redirect_uri") // Optional: frontend redirect after auth
	flow := c.DefaultQuery("flow", "login")
	plan := c.Query("plan")

	var oauthConfig *oauth2.Config

	switch provider {
	case "google":
		if !h.cfg.HasGoogleOAuth() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider_not_configured", "message": "Google OAuth is not configured"})
			return
		}
		oauthConfig = &oauth2.Config{
			ClientID:     h.cfg.GoogleClientID,
			ClientSecret: h.cfg.GoogleClientSecret,
			RedirectURL:  h.cfg.AppURL + "/api/v1/auth/social/callback",
			Scopes:       []string{"email", "profile"},
			Endpoint:     google.Endpoint,
		}
	case "github":
		if !h.cfg.HasGitHubOAuth() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider_not_configured", "message": "GitHub OAuth is not configured"})
			return
		}
		oauthConfig = &oauth2.Config{
			ClientID:     h.cfg.GitHubClientID,
			ClientSecret: h.cfg.GitHubClientSecret,
			RedirectURL:  h.cfg.AppURL + "/api/v1/auth/social/callback",
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_provider", "message": "Unsupported OAuth provider"})
		return
	}

	// Generate state token
	stateBytes := make([]byte, 32)
	rand.Read(stateBytes)
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store state in database
	oauthState := models.OAuthState{
		State:     state,
		Provider:  provider,
		Plan:      plan,
		Flow:      flow,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	h.db.Create(&oauthState)

	// Generate auth URL
	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// HandleOAuthCallback handles the OAuth callback
// POST /api/v1/auth/social/callback
func (h *AuthHandler) HandleOAuthCallback(c *gin.Context) {
	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Code and state are required"})
		return
	}

	// Verify state
	var oauthState models.OAuthState
	if err := h.db.Where("state = ?", req.State).First(&oauthState).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_state", "message": "Invalid or expired state"})
		return
	}

	// Delete used state
	h.db.Delete(&oauthState)

	// Check expiry
	if time.Now().After(oauthState.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state_expired", "message": "OAuth state has expired"})
		return
	}

	// Exchange code for user info
	email, name, picture, err := h.exchangeOAuthCode(oauthState.Provider, req.Code)
	if err != nil {
		log.Printf("OAuth exchange failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth_failed", "message": "Failed to authenticate with provider"})
		return
	}

	// Find or create user
	var user models.User
	result := h.db.Where("email = ?", email).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		user = models.User{
			Email:         email,
			Name:          name,
			Picture:       picture,
			AuthProvider:  oauthState.Provider,
			EmailVerified: true,
			LastLogin:     time.Now(),
		}
		if oauthState.Plan != "" {
			user.SelectedPlanTier = models.PlanTier(oauthState.Plan)
		}
		if err := h.db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create account"})
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Database error"})
		return
	} else {
		// Update existing user
		user.LastLogin = time.Now()
		user.Name = name
		user.Picture = picture
		if user.AuthProvider == "" {
			user.AuthProvider = oauthState.Provider
			user.EmailVerified = true
		}
		h.db.Save(&user)
	}

	// Generate JWT
	token, err := h.generateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":       token,
		"user":               userResponse(&user),
		"needs_tenant_setup": user.AdminOfTenantID == nil,
		"flow":               oauthState.Flow,
	})
}

// ============================================================================
// Email/Password Auth
// ============================================================================

// Register handles email/password registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Name     string `json:"name"`
		Plan     string `json:"plan"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Invalid email or password"})
		return
	}

	// Check if email exists
	var existing models.User
	if err := h.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email_exists", "message": "An account with this email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to process password"})
		return
	}

	// Generate verification token
	verifyToken := generateRandomToken(32)
	expiry := time.Now().Add(24 * time.Hour)

	user := models.User{
		Email:         req.Email,
		Name:          req.Name,
		AuthProvider:  "local",
		EmailVerified: false,
		PasswordHash:  string(hashedPassword),
		VerifyToken:   verifyToken,
		VerifyExpiry:  &expiry,
	}
	if req.Plan != "" {
		user.SelectedPlanTier = models.PlanTier(req.Plan)
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create account"})
		return
	}

	// TODO: Send verification email

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Account created. Please check your email to verify your account.",
		"verify_token": verifyToken, // In production, only send via email
	})
}

// VerifyEmail verifies email address
// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Token is required"})
		return
	}

	var user models.User
	if err := h.db.Where("verify_token = ?", req.Token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_token", "message": "Invalid verification token"})
		return
	}

	if user.IsVerifyTokenExpired() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token_expired", "message": "Verification token has expired"})
		return
	}

	user.EmailVerified = true
	user.VerifyToken = ""
	user.VerifyExpiry = nil
	user.LastLogin = time.Now()
	h.db.Save(&user)

	token, _ := h.generateToken(&user)

	c.JSON(http.StatusOK, gin.H{
		"access_token":       token,
		"user":               userResponse(&user),
		"needs_tenant_setup": user.AdminOfTenantID == nil,
	})
}

// Login handles email/password login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Invalid credentials"})
		return
	}

	var user models.User
	if err := h.db.Where("email = ? AND auth_provider = ?", req.Email, "local").First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials", "message": "Invalid email or password"})
		return
	}

	if !user.EmailVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "email_not_verified", "message": "Please verify your email first"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials", "message": "Invalid email or password"})
		return
	}

	user.LastLogin = time.Now()
	h.db.Save(&user)

	token, _ := h.generateToken(&user)

	c.JSON(http.StatusOK, gin.H{
		"access_token":       token,
		"user":               userResponse(&user),
		"needs_tenant_setup": user.AdminOfTenantID == nil,
	})
}

// ForgotPassword initiates password reset
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Invalid email"})
		return
	}

	var user models.User
	if err := h.db.Where("email = ? AND auth_provider = ?", req.Email, "local").First(&user).Error; err != nil {
		// Don't reveal if email exists
		c.JSON(http.StatusOK, gin.H{"message": "If an account exists, a reset link has been sent."})
		return
	}

	resetToken := generateRandomToken(32)
	expiry := time.Now().Add(1 * time.Hour)
	user.ResetToken = resetToken
	user.ResetExpiry = &expiry
	h.db.Save(&user)

	// TODO: Send reset email

	c.JSON(http.StatusOK, gin.H{
		"message":     "If an account exists, a reset link has been sent.",
		"reset_token": resetToken, // In production, only send via email
	})
}

// ResetPassword resets the password
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Invalid request"})
		return
	}

	var user models.User
	if err := h.db.Where("reset_token = ?", req.Token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_token", "message": "Invalid reset token"})
		return
	}

	if user.IsResetTokenExpired() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token_expired", "message": "Reset token has expired"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user.PasswordHash = string(hashedPassword)
	user.ResetToken = ""
	user.ResetExpiry = nil
	h.db.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

// GetCurrentUser returns the current user
// GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, userResponse(&user))
}

// ============================================================================
// Helpers
// ============================================================================

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":            user.ID.String(),
		"email":          user.Email,
		"name":           user.Name,
		"type":           "platform",
		"email_verified": user.EmailVerified,
		"is_tenant_admin": user.IsTenantAdmin,
		"iat":            time.Now().Unix(),
		"exp":            time.Now().Add(24 * time.Hour).Unix(),
	}

	if user.AdminOfTenantID != nil {
		claims["tenant_id"] = user.AdminOfTenantID.String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.cfg.GetJWTSecret())
}

func (h *AuthHandler) exchangeOAuthCode(provider, code string) (email, name, picture string, err error) {
	var oauthConfig *oauth2.Config

	switch provider {
	case "google":
		oauthConfig = &oauth2.Config{
			ClientID:     h.cfg.GoogleClientID,
			ClientSecret: h.cfg.GoogleClientSecret,
			RedirectURL:  h.cfg.AppURL + "/api/v1/auth/social/callback",
			Scopes:       []string{"email", "profile"},
			Endpoint:     google.Endpoint,
		}
	case "github":
		oauthConfig = &oauth2.Config{
			ClientID:     h.cfg.GitHubClientID,
			ClientSecret: h.cfg.GitHubClientSecret,
			RedirectURL:  h.cfg.AppURL + "/api/v1/auth/social/callback",
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	default:
		return "", "", "", fmt.Errorf("unsupported provider: %s", provider)
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return "", "", "", err
	}

	client := oauthConfig.Client(context.Background(), token)

	switch provider {
	case "google":
		return h.getGoogleUserInfo(client)
	case "github":
		return h.getGitHubUserInfo(client)
	}

	return "", "", "", fmt.Errorf("unsupported provider")
}

func (h *AuthHandler) getGoogleUserInfo(client *http.Client) (email, name, picture string, err error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var userInfo struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.Unmarshal(body, &userInfo); err != nil {
		return "", "", "", err
	}

	return userInfo.Email, userInfo.Name, userInfo.Picture, nil
}

func (h *AuthHandler) getGitHubUserInfo(client *http.Client) (email, name, picture string, err error) {
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var userInfo struct {
		Email     string `json:"email"`
		Name      string `json:"name"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &userInfo); err != nil {
		return "", "", "", err
	}

	// Fetch email if not public
	if userInfo.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			emailBody, _ := io.ReadAll(emailResp.Body)

			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.Unmarshal(emailBody, &emails) == nil {
				for _, e := range emails {
					if e.Primary {
						userInfo.Email = e.Email
						break
					}
				}
			}
		}
	}

	if userInfo.Name == "" {
		userInfo.Name = userInfo.Login
	}

	return userInfo.Email, userInfo.Name, userInfo.AvatarURL, nil
}

func userResponse(user *models.User) gin.H {
	resp := gin.H{
		"id":              user.ID,
		"email":           user.Email,
		"name":            user.Name,
		"picture":         user.Picture,
		"auth_provider":   user.AuthProvider,
		"email_verified":  user.EmailVerified,
		"is_tenant_admin": user.IsTenantAdmin,
		"created_at":      user.CreatedAt,
	}

	if user.AdminOfTenantID != nil {
		resp["tenant_id"] = user.AdminOfTenantID
	}

	if user.SelectedPlanTier != "" {
		resp["selected_plan"] = user.SelectedPlanTier
	}

	return resp
}

func generateRandomToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
