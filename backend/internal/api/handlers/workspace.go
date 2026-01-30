package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/saas-starter-kit/backend/internal/config"
	"github.com/yourusername/saas-starter-kit/backend/internal/models"
	"gorm.io/gorm"
)

type WorkspaceHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewWorkspaceHandler(db *gorm.DB, cfg *config.Config) *WorkspaceHandler {
	return &WorkspaceHandler{db: db, cfg: cfg}
}

// List returns all workspaces for the current tenant
// GET /api/v1/workspaces
func (h *WorkspaceHandler) List(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")

	var workspaces []models.Workspace
	if err := h.db.Where("tenant_id = ?", tenantID).Order("created_at ASC").Find(&workspaces).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to fetch workspaces"})
		return
	}

	result := make([]gin.H, len(workspaces))
	for i, ws := range workspaces {
		result[i] = workspaceResponse(&ws)
	}

	c.JSON(http.StatusOK, gin.H{"workspaces": result})
}

// Create creates a new workspace
// POST /api/v1/workspaces
func (h *WorkspaceHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Slug string `json:"slug"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Workspace name is required"})
		return
	}

	tenantID, _ := c.Get("tenant_id")
	userID, _ := c.Get("user_id")

	// Parse tenant ID
	tenantUUID, err := uuid.Parse(tenantID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_tenant", "message": "Invalid tenant ID"})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user", "message": "Invalid user ID"})
		return
	}

	// Check workspace limit
	var tenant models.Tenant
	if err := h.db.Preload("Subscription.Plan").First(&tenant, "id = ?", tenantUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant_not_found", "message": "Tenant not found"})
		return
	}

	if tenant.Subscription != nil && tenant.Subscription.Plan.MaxWorkspaces > 0 {
		var count int64
		h.db.Model(&models.Workspace{}).Where("tenant_id = ?", tenantUUID).Count(&count)
		if int(count) >= tenant.Subscription.Plan.MaxWorkspaces {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "workspace_limit_reached",
				"message": "You have reached the maximum number of workspaces for your plan",
				"limit":   tenant.Subscription.Plan.MaxWorkspaces,
			})
			return
		}
	}

	// Generate or validate slug
	slug := req.Slug
	if slug == "" {
		slug = strings.ToLower(regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(req.Name, "-"))
		slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
		slug = strings.Trim(slug, "-")
	}

	if len(slug) < 2 {
		slug = slug + "-workspace"
	}

	// Check if slug exists within tenant
	var existingCount int64
	h.db.Model(&models.Workspace{}).Where("tenant_id = ? AND slug = ?", tenantUUID, slug).Count(&existingCount)
	if existingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "slug_exists", "message": "A workspace with this slug already exists"})
		return
	}

	// Start transaction
	tx := h.db.Begin()

	// Create workspace
	workspace := models.Workspace{
		TenantID:    tenantUUID,
		Slug:        slug,
		DisplayName: req.Name,
		IsDefault:   false,
	}

	if err := tx.Create(&workspace).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create workspace"})
		return
	}

	// Add creator as workspace admin
	membership := models.Membership{
		UserID:      userUUID,
		WorkspaceID: workspace.ID,
		Role:        "admin",
	}

	if err := tx.Create(&membership).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create membership"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Workspace created successfully",
		"workspace": workspaceResponse(&workspace),
	})
}

// Get returns a specific workspace
// GET /api/v1/workspaces/:id
func (h *WorkspaceHandler) Get(c *gin.Context) {
	workspaceID := c.Param("id")
	tenantID, _ := c.Get("tenant_id")
	userID, _ := c.Get("user_id")

	// Parse workspace ID (can be UUID or slug)
	var workspace models.Workspace
	var query *gorm.DB

	if _, err := uuid.Parse(workspaceID); err == nil {
		query = h.db.Where("id = ? AND tenant_id = ?", workspaceID, tenantID)
	} else {
		query = h.db.Where("slug = ? AND tenant_id = ?", workspaceID, tenantID)
	}

	if err := query.First(&workspace).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Workspace not found"})
		return
	}

	// Check if user has access to this workspace
	var membership models.Membership
	if err := h.db.Where("user_id = ? AND workspace_id = ?", userID, workspace.ID).First(&membership).Error; err != nil {
		// Check if user is tenant admin
		var user models.User
		h.db.First(&user, "id = ?", userID)
		if user.AdminOfTenantID == nil || *user.AdminOfTenantID != workspace.TenantID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access_denied", "message": "You don't have access to this workspace"})
			return
		}
	}

	response := workspaceResponse(&workspace)
	if membership.ID != uuid.Nil {
		response["role"] = membership.Role
	} else {
		response["role"] = "admin" // Tenant admin
	}

	c.JSON(http.StatusOK, response)
}

// Delete deletes a workspace
// DELETE /api/v1/workspaces/:id
func (h *WorkspaceHandler) Delete(c *gin.Context) {
	workspaceID := c.Param("id")
	tenantID, _ := c.Get("tenant_id")
	userID, _ := c.Get("user_id")

	var workspace models.Workspace
	if err := h.db.Where("id = ? AND tenant_id = ?", workspaceID, tenantID).First(&workspace).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Workspace not found"})
		return
	}

	// Cannot delete default workspace
	if workspace.IsDefault {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot_delete_default", "message": "Cannot delete the default workspace"})
		return
	}

	// Check if user is workspace admin or tenant admin
	var membership models.Membership
	if err := h.db.Where("user_id = ? AND workspace_id = ? AND role = ?", userID, workspace.ID, "admin").First(&membership).Error; err != nil {
		// Check if tenant admin
		var user models.User
		h.db.First(&user, "id = ?", userID)
		if user.AdminOfTenantID == nil || *user.AdminOfTenantID != workspace.TenantID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access_denied", "message": "Only workspace or tenant admins can delete workspaces"})
			return
		}
	}

	// Delete workspace (cascades to memberships)
	if err := h.db.Delete(&workspace).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to delete workspace"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workspace deleted successfully"})
}

// AddMember adds a user to a workspace
// POST /api/v1/workspaces/:id/members
func (h *WorkspaceHandler) AddMember(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Valid email is required"})
		return
	}

	workspaceID := c.Param("id")
	tenantID, _ := c.Get("tenant_id")

	var workspace models.Workspace
	if err := h.db.Where("id = ? AND tenant_id = ?", workspaceID, tenantID).First(&workspace).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Workspace not found"})
		return
	}

	// Find user by email
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found", "message": "User not found with this email"})
		return
	}

	// Check if already a member
	var existingMembership models.Membership
	if err := h.db.Where("user_id = ? AND workspace_id = ?", user.ID, workspace.ID).First(&existingMembership).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already_member", "message": "User is already a member of this workspace"})
		return
	}

	// Set default role
	role := req.Role
	if role == "" {
		role = "member"
	}
	if role != "admin" && role != "member" && role != "viewer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_role", "message": "Role must be admin, member, or viewer"})
		return
	}

	// Create membership
	membership := models.Membership{
		UserID:      user.ID,
		WorkspaceID: workspace.ID,
		Role:        role,
	}

	if err := h.db.Create(&membership).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Member added successfully",
		"member": gin.H{
			"user_id":      user.ID,
			"email":        user.Email,
			"name":         user.Name,
			"role":         membership.Role,
			"workspace_id": workspace.ID,
		},
	})
}

// ListMembers returns all members of a workspace
// GET /api/v1/workspaces/:id/members
func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
	workspaceID := c.Param("id")
	tenantID, _ := c.Get("tenant_id")

	var workspace models.Workspace
	if err := h.db.Where("id = ? AND tenant_id = ?", workspaceID, tenantID).First(&workspace).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Workspace not found"})
		return
	}

	var memberships []models.Membership
	if err := h.db.Preload("User").Where("workspace_id = ?", workspace.ID).Find(&memberships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to fetch members"})
		return
	}

	members := make([]gin.H, len(memberships))
	for i, m := range memberships {
		members[i] = gin.H{
			"user_id":    m.UserID,
			"email":      m.User.Email,
			"name":       m.User.Name,
			"picture":    m.User.Picture,
			"role":       m.Role,
			"created_at": m.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}
