package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/saas-starter-kit/backend/internal/config"
	"github.com/yourusername/saas-starter-kit/backend/internal/hierarchy"
	"gorm.io/gorm"
)

// ContainerHandler handles CRUD operations for any level in the hierarchy
type ContainerHandler struct {
	db         *gorm.DB
	cfg        *config.Config
	hierarchy  *hierarchy.Config
	repository *hierarchy.Repository
}

// NewContainerHandler creates a new container handler
func NewContainerHandler(db *gorm.DB, cfg *config.Config, h *hierarchy.Config) *ContainerHandler {
	return &ContainerHandler{
		db:         db,
		cfg:        cfg,
		hierarchy:  h,
		repository: hierarchy.NewRepository(db, h),
	}
}

// ListContainers lists containers at a given level
// GET /api/v1/{level_url_path}
func (h *ContainerHandler) ListContainers(c *gin.Context) {
	level := c.Param("level")
	levelConfig := h.hierarchy.GetLevel(level)
	if levelConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_level", "message": "Unknown hierarchy level"})
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuid.Parse(userID.(string))

	// For root level, check if user is root admin
	if levelConfig.IsRoot {
		rootID, exists := c.Get("root_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "no_root_access", "message": "User has no organization"})
			return
		}
		rootUUID, _ := uuid.Parse(rootID.(string))
		container, err := h.repository.GetContainer(rootUUID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Organization not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			levelConfig.Name: containerResponse(container, levelConfig),
		})
		return
	}

	// For non-root levels, list containers user has access to
	containers, err := h.repository.GetUserContainers(userUUID, level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to fetch containers"})
		return
	}

	result := make([]gin.H, len(containers))
	for i, container := range containers {
		result[i] = containerResponse(&container, levelConfig)
	}

	c.JSON(http.StatusOK, gin.H{levelConfig.Plural: result})
}

// CreateContainer creates a new container
// POST /api/v1/{level_url_path}
func (h *ContainerHandler) CreateContainer(c *gin.Context) {
	level := c.Param("level")
	levelConfig := h.hierarchy.GetLevel(level)
	if levelConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_level", "message": "Unknown hierarchy level"})
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required"`
		Slug     string `json:"slug"`
		ParentID string `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Name is required"})
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuid.Parse(userID.(string))

	// Generate slug if not provided
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	// Determine parent
	var parentID *uuid.UUID
	if levelConfig.IsRoot {
		// Root level has no parent
		parentID = nil
	} else {
		// Non-root levels require a parent
		if req.ParentID == "" {
			// Use parent from context (e.g., X-Parent-ID header or root_id)
			parentLevel := h.hierarchy.GetParentLevel(level)
			if parentLevel != nil && parentLevel.IsRoot {
				rootID, _ := c.Get("root_id")
				if rootID != nil {
					id, _ := uuid.Parse(rootID.(string))
					parentID = &id
				}
			}
		} else {
			id, _ := uuid.Parse(req.ParentID)
			parentID = &id
		}

		if parentID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parent_required", "message": "Parent container is required"})
			return
		}
	}

	// Create container
	container, err := h.repository.CreateContainer(level, slug, req.Name, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to create container"})
		return
	}

	// Add creator as admin
	if err := h.repository.AddMember(userUUID, container.ID, "admin"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to add membership"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        levelConfig.DisplayName + " created successfully",
		levelConfig.Name: containerResponse(container, levelConfig),
	})
}

// GetContainer retrieves a specific container
// GET /api/v1/{level_url_path}/:id
func (h *ContainerHandler) GetContainer(c *gin.Context) {
	level := c.Param("level")
	levelConfig := h.hierarchy.GetLevel(level)
	if levelConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_level", "message": "Unknown hierarchy level"})
		return
	}

	containerID := c.Param("id")

	// Try to parse as UUID, otherwise treat as slug
	var container *hierarchy.ResourceContainer
	var err error

	if id, parseErr := uuid.Parse(containerID); parseErr == nil {
		container, err = h.repository.GetContainer(id)
	} else {
		// Lookup by slug within parent context
		var parentID *uuid.UUID
		if !levelConfig.IsRoot {
			rootID, _ := c.Get("root_id")
			if rootID != nil {
				id, _ := uuid.Parse(rootID.(string))
				parentID = &id
			}
		}
		container, err = h.repository.GetContainerBySlug(level, containerID, parentID)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": levelConfig.DisplayName + " not found"})
		return
	}

	c.JSON(http.StatusOK, containerResponse(container, levelConfig))
}

// DeleteContainer deletes a container
// DELETE /api/v1/{level_url_path}/:id
func (h *ContainerHandler) DeleteContainer(c *gin.Context) {
	level := c.Param("level")
	levelConfig := h.hierarchy.GetLevel(level)
	if levelConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_level", "message": "Unknown hierarchy level"})
		return
	}

	containerID := c.Param("id")
	id, err := uuid.Parse(containerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid container ID"})
		return
	}

	container, err := h.repository.GetContainer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": levelConfig.DisplayName + " not found"})
		return
	}

	// Cannot delete root container
	if levelConfig.IsRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot_delete_root", "message": "Cannot delete root organization"})
		return
	}

	// Delete container (cascades to children and memberships)
	if err := h.db.Delete(container).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to delete container"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": levelConfig.DisplayName + " deleted successfully"})
}

// ListMembers lists members of a container
// GET /api/v1/{level_url_path}/:id/members
func (h *ContainerHandler) ListMembers(c *gin.Context) {
	level := c.Param("level")
	levelConfig := h.hierarchy.GetLevel(level)
	if levelConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_level", "message": "Unknown hierarchy level"})
		return
	}

	containerID := c.Param("id")
	id, err := uuid.Parse(containerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid container ID"})
		return
	}

	memberships, err := h.repository.ListMembers(id)
	if err != nil {
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

// AddMember adds a member to a container
// POST /api/v1/{level_url_path}/:id/members
func (h *ContainerHandler) AddMember(c *gin.Context) {
	level := c.Param("level")
	levelConfig := h.hierarchy.GetLevel(level)
	if levelConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid_level", "message": "Unknown hierarchy level"})
		return
	}

	containerID := c.Param("id")
	id, err := uuid.Parse(containerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "Invalid container ID"})
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Valid email is required"})
		return
	}

	// Find user by email
	var user hierarchy.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found", "message": "User not found with this email"})
		return
	}

	// Check if already a member
	if _, err := h.repository.GetMembership(user.ID, id); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already_member", "message": "User is already a member"})
		return
	}

	// Validate role
	role := req.Role
	if role == "" {
		role = "member"
	}
	validRole := false
	for _, r := range levelConfig.Roles {
		if r == role {
			validRole = true
			break
		}
	}
	if !validRole {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "invalid_role",
			"message":     "Invalid role for this level",
			"valid_roles": levelConfig.Roles,
		})
		return
	}

	// Add member
	if err := h.repository.AddMember(user.ID, id, role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Member added successfully",
		"member": gin.H{
			"user_id": user.ID,
			"email":   user.Email,
			"name":    user.Name,
			"role":    role,
		},
	})
}

// GetHierarchyConfig returns the hierarchy configuration
// GET /api/v1/hierarchy
func (h *ContainerHandler) GetHierarchyConfig(c *gin.Context) {
	levels := make([]gin.H, len(h.hierarchy.Levels))
	for i, level := range h.hierarchy.Levels {
		levels[i] = gin.H{
			"name":         level.Name,
			"display_name": level.DisplayName,
			"plural":       level.Plural,
			"url_path":     level.URLPath,
			"roles":        level.Roles,
			"is_root":      level.IsRoot,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"root_level": h.hierarchy.RootLevel,
		"leaf_level": h.hierarchy.LeafLevel,
		"depth":      h.hierarchy.Depth(),
		"levels":     levels,
	})
}

// Helper functions

func containerResponse(container *hierarchy.ResourceContainer, levelConfig *hierarchy.Level) gin.H {
	return gin.H{
		"id":           container.ID,
		"level":        container.Level,
		"slug":         container.Slug,
		"display_name": container.DisplayName,
		"parent_id":    container.ParentID,
		"root_id":      container.RootID,
		"depth":        container.Depth,
		"is_active":    container.IsActive,
		"created_at":   container.CreatedAt,
		// Include level metadata
		"_level_config": gin.H{
			"display_name": levelConfig.DisplayName,
			"plural":       levelConfig.Plural,
			"roles":        levelConfig.Roles,
		},
	}
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 50 {
		slug = slug[:50]
	}
	return slug
}
