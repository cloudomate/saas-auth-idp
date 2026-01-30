package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourusername/saas-starter-kit/backend/internal/config"
	"github.com/yourusername/saas-starter-kit/backend/internal/models"
	"gorm.io/gorm"
)

type TenantHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewTenantHandler(db *gorm.DB, cfg *config.Config) *TenantHandler {
	return &TenantHandler{db: db, cfg: cfg}
}

// GetCurrentTenant returns the current user's tenant
// GET /api/v1/tenant
func (h *TenantHandler) GetCurrentTenant(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found", "message": "User not found"})
		return
	}

	if user.AdminOfTenantID == nil {
		c.JSON(http.StatusOK, gin.H{
			"tenant":             nil,
			"needs_tenant_setup": true,
			"selected_plan":      user.SelectedPlanTier,
		})
		return
	}

	var tenant models.Tenant
	if err := h.db.Preload("Subscription.Plan").First(&tenant, "id = ?", user.AdminOfTenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant_not_found", "message": "Tenant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant":             tenantResponse(&tenant),
		"needs_tenant_setup": false,
	})
}

// ListPlans returns available subscription plans
// GET /api/v1/tenant/plans
func (h *TenantHandler) ListPlans(c *gin.Context) {
	var plans []models.Plan
	if err := h.db.Where("is_active = ?", true).Order("monthly_price_cents ASC").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to fetch plans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

// SelectPlan saves the user's plan selection
// POST /api/v1/tenant/select-plan
func (h *TenantHandler) SelectPlan(c *gin.Context) {
	var req struct {
		Plan string `json:"plan" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Plan is required"})
		return
	}

	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found", "message": "User not found"})
		return
	}

	// Validate plan tier
	planTier := models.PlanTier(req.Plan)
	if planTier != models.PlanTierBasic && planTier != models.PlanTierAdvanced && planTier != models.PlanTierEnterprise {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_plan", "message": "Invalid plan tier"})
		return
	}

	user.SelectedPlanTier = planTier
	h.db.Save(&user)

	// For Basic plan, auto-create tenant
	if planTier == models.PlanTierBasic {
		tenant, token, err := h.autoCreateTenant(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create tenant"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Plan selected and tenant created",
			"tenant":       tenantResponse(tenant),
			"access_token": token,
			"redirect_to":  "/dashboard",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Plan selected",
		"redirect_to": "/setup/organization",
	})
}

// CheckSlug checks if a tenant slug is available
// GET /api/v1/tenant/check-slug?slug=xxx
func (h *TenantHandler) CheckSlug(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Slug is required"})
		return
	}

	// Validate slug format
	slugRegex := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if len(slug) < 3 || len(slug) > 50 || !slugRegex.MatchString(slug) {
		c.JSON(http.StatusOK, gin.H{"available": false, "reason": "invalid_format"})
		return
	}

	// Check reserved slugs
	reserved := []string{"admin", "api", "www", "app", "dashboard", "settings", "login", "signup", "auth"}
	for _, r := range reserved {
		if slug == r {
			c.JSON(http.StatusOK, gin.H{"available": false, "reason": "reserved"})
			return
		}
	}

	// Check if slug exists
	var count int64
	h.db.Model(&models.Tenant{}).Where("slug = ?", slug).Count(&count)

	c.JSON(http.StatusOK, gin.H{"available": count == 0})
}

// SetupOrganization creates a new tenant for the user
// POST /api/v1/tenant/setup
func (h *TenantHandler) SetupOrganization(c *gin.Context) {
	var req struct {
		OrgName     string `json:"org_name" binding:"required"`
		OrgSlug     string `json:"org_slug" binding:"required"`
		EmailDomain string `json:"email_domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Organization name and slug are required"})
		return
	}

	userID, _ := c.Get("user_id")

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found", "message": "User not found"})
		return
	}

	// Check if user already has a tenant
	if user.AdminOfTenantID != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already_has_tenant", "message": "User already has an organization"})
		return
	}

	// Validate and normalize slug
	slug := strings.ToLower(strings.TrimSpace(req.OrgSlug))
	slugRegex := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if len(slug) < 3 || len(slug) > 50 || !slugRegex.MatchString(slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_slug", "message": "Invalid slug format"})
		return
	}

	// Check if slug is taken
	var count int64
	h.db.Model(&models.Tenant{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "slug_exists", "message": "This organization URL is already taken"})
		return
	}

	// Start transaction
	tx := h.db.Begin()

	// Create tenant
	tenant := models.Tenant{
		Slug:        slug,
		DisplayName: req.OrgName,
		AdminUserID: &user.ID,
		IsActive:    true,
	}

	if err := tx.Create(&tenant).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create organization"})
		return
	}

	// Get the plan based on user's selection
	planTier := user.SelectedPlanTier
	if planTier == "" {
		planTier = models.PlanTierBasic
	}

	var plan models.Plan
	if err := tx.Where("tier = ?", planTier).First(&plan).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to find plan"})
		return
	}

	// Create subscription
	subscription := models.Subscription{
		TenantID:           tenant.ID,
		PlanID:             plan.ID,
		Status:             "active",
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(0, 1, 0), // 1 month from now
	}

	if err := tx.Create(&subscription).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create subscription"})
		return
	}

	// Create default workspace
	workspace := models.Workspace{
		TenantID:    tenant.ID,
		Slug:        "default",
		DisplayName: "Default Workspace",
		IsDefault:   true,
	}

	if err := tx.Create(&workspace).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create workspace"})
		return
	}

	// Add user as workspace admin
	membership := models.Membership{
		UserID:      user.ID,
		WorkspaceID: workspace.ID,
		Role:        "admin",
	}

	if err := tx.Create(&membership).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create membership"})
		return
	}

	// Update user as tenant admin
	user.IsTenantAdmin = true
	user.AdminOfTenantID = &tenant.ID
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to update user"})
		return
	}

	tx.Commit()

	// Generate new token with tenant_id
	token, _ := h.generateTenantToken(&user, &tenant)

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Organization created successfully",
		"tenant":       tenantResponse(&tenant),
		"workspace":    workspaceResponse(&workspace),
		"access_token": token,
	})
}

// ============================================================================
// Helpers
// ============================================================================

func (h *TenantHandler) autoCreateTenant(user *models.User) (*models.Tenant, string, error) {
	tx := h.db.Begin()

	// Generate slug from email
	emailPrefix := strings.Split(user.Email, "@")[0]
	slug := strings.ToLower(regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(emailPrefix, "-"))
	if len(slug) < 3 {
		slug = slug + "-workspace"
	}

	// Ensure unique slug
	baseSlug := slug
	for i := 1; ; i++ {
		var count int64
		h.db.Model(&models.Tenant{}).Where("slug = ?", slug).Count(&count)
		if count == 0 {
			break
		}
		slug = baseSlug + "-" + string(rune('0'+i))
		if i > 9 {
			slug = baseSlug + "-" + uuid.New().String()[:8]
			break
		}
	}

	// Create tenant
	tenant := models.Tenant{
		Slug:        slug,
		DisplayName: user.Name + "'s Workspace",
		AdminUserID: &user.ID,
		IsActive:    true,
	}

	if err := tx.Create(&tenant).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	// Get Basic plan
	var plan models.Plan
	if err := tx.Where("tier = ?", models.PlanTierBasic).First(&plan).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	// Create subscription
	subscription := models.Subscription{
		TenantID:           tenant.ID,
		PlanID:             plan.ID,
		Status:             "active",
		CurrentPeriodStart: time.Now(),
		CurrentPeriodEnd:   time.Now().AddDate(1, 0, 0), // 1 year for free plan
	}

	if err := tx.Create(&subscription).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	// Create default workspace
	workspace := models.Workspace{
		TenantID:    tenant.ID,
		Slug:        "default",
		DisplayName: "My Workspace",
		IsDefault:   true,
	}

	if err := tx.Create(&workspace).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	// Add user as workspace admin
	membership := models.Membership{
		UserID:      user.ID,
		WorkspaceID: workspace.ID,
		Role:        "admin",
	}

	if err := tx.Create(&membership).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	// Update user
	user.IsTenantAdmin = true
	user.AdminOfTenantID = &tenant.ID
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return nil, "", err
	}

	tx.Commit()

	// Generate token
	token, _ := h.generateTenantToken(user, &tenant)

	return &tenant, token, nil
}

func (h *TenantHandler) generateTenantToken(user *models.User, tenant *models.Tenant) (string, error) {
	claims := jwt.MapClaims{
		"sub":             user.ID.String(),
		"email":           user.Email,
		"name":            user.Name,
		"type":            "platform",
		"email_verified":  user.EmailVerified,
		"is_tenant_admin": true,
		"tenant_id":       tenant.ID.String(),
		"iat":             time.Now().Unix(),
		"exp":             time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.cfg.GetJWTSecret())
}

func tenantResponse(tenant *models.Tenant) gin.H {
	resp := gin.H{
		"id":             tenant.ID,
		"slug":           tenant.Slug,
		"display_name":   tenant.DisplayName,
		"is_active":      tenant.IsActive,
		"sso_configured": tenant.SSOConfigured,
		"created_at":     tenant.CreatedAt,
	}

	if tenant.Subscription != nil {
		resp["subscription"] = gin.H{
			"id":     tenant.Subscription.ID,
			"status": tenant.Subscription.Status,
			"plan":   tenant.Subscription.Plan,
		}
	}

	return resp
}

func workspaceResponse(ws *models.Workspace) gin.H {
	return gin.H{
		"id":           ws.ID,
		"tenant_id":    ws.TenantID,
		"slug":         ws.Slug,
		"display_name": ws.DisplayName,
		"is_default":   ws.IsDefault,
		"created_at":   ws.CreatedAt,
	}
}
