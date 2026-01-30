package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sample-api/internal/authz"
	"github.com/yourusername/sample-api/internal/middleware"
	"github.com/yourusername/sample-api/internal/store"
)

// ProjectHandler handles project operations
// This demonstrates ABAC (Attribute-Based Access Control)
//
// ABAC Pattern:
// - Access is determined by attributes of the user, resource, and environment
// - Attributes considered:
//   - User: roles, department, clearance level
//   - Resource: environment (prod/staging/dev), status, tags
//   - Context: time of day, IP address, etc.
//
// Example Policies:
// - Only admins can deploy to production
// - Developers can deploy to staging/development
// - Archived projects are read-only
// - Production projects require approval for changes
type ProjectHandler struct {
	store *store.MemoryStore
	fga   *authz.OpenFGAClient
}

func NewProjectHandler(s *store.MemoryStore, fga *authz.OpenFGAClient) *ProjectHandler {
	return &ProjectHandler{store: s, fga: fga}
}

// List returns all projects in the workspace
// GET /api/v1/projects
func (h *ProjectHandler) List(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)

	// ABAC: Filter by environment if user doesn't have full access
	env := c.Query("environment")

	var projects []*store.Project
	if env != "" {
		projects = h.store.ListProjectsByEnvironment(userCtx.WorkspaceID, env)
	} else {
		projects = h.store.ListProjects(userCtx.WorkspaceID)
	}

	// Enrich with permissions based on ABAC policies
	type ProjectWithPermissions struct {
		*store.Project
		Permissions map[string]bool `json:"permissions"`
	}

	result := make([]ProjectWithPermissions, 0, len(projects))
	for _, proj := range projects {
		permissions := h.evaluateABACPolicies(userCtx, proj)
		result = append(result, ProjectWithPermissions{
			Project:     proj,
			Permissions: permissions,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": result,
		"total":    len(result),
		"policies": h.getActivePolicies(),
	})
}

// Create creates a new project
// POST /api/v1/projects
func (h *ProjectHandler) Create(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Environment string   `json:"environment"` // production, staging, development
		Tags        []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	if req.Environment == "" {
		req.Environment = "development"
	}

	// ABAC Policy: Only admins can create production projects
	if req.Environment == "production" && !h.isAdmin(userCtx) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "policy_violation",
			"message": "Only administrators can create production projects",
			"policy":  "create_production_project",
		})
		return
	}

	proj := &store.Project{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		WorkspaceID: userCtx.WorkspaceID,
		OwnerID:     userCtx.UserID,
		Environment: req.Environment,
		Status:      "active",
		Tags:        req.Tags,
	}

	if err := h.store.CreateProject(proj); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"project":     proj,
		"permissions": h.evaluateABACPolicies(userCtx, proj),
	})
}

// Get returns a specific project
// GET /api/v1/projects/:id
func (h *ProjectHandler) Get(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)
	projID := c.Param("id")

	proj, err := h.store.GetProject(projID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	// Workspace isolation
	if proj.WorkspaceID != userCtx.WorkspaceID && !userCtx.IsPlatformAdmin {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	permissions := h.evaluateABACPolicies(userCtx, proj)

	c.JSON(http.StatusOK, gin.H{
		"project":     proj,
		"permissions": permissions,
	})
}

// Update updates a project
// PUT /api/v1/projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)
	projID := c.Param("id")

	proj, err := h.store.GetProject(projID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	permissions := h.evaluateABACPolicies(userCtx, proj)

	// ABAC Policy: Check write permission
	if !permissions["can_write"] {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "policy_violation",
			"message": "You don't have permission to modify this project",
			"reason":  h.getWriteDenialReason(userCtx, proj),
		})
		return
	}

	var req struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Environment *string  `json:"environment"`
		Status      *string  `json:"status"`
		Tags        []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// ABAC Policy: Environment change restrictions
	if req.Environment != nil && *req.Environment != proj.Environment {
		// Can't move to production without admin role
		if *req.Environment == "production" && !h.isAdmin(userCtx) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "policy_violation",
				"message": "Only administrators can promote projects to production",
				"policy":  "promote_to_production",
			})
			return
		}

		// Can't demote production without admin role
		if proj.Environment == "production" && !h.isAdmin(userCtx) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "policy_violation",
				"message": "Only administrators can modify production projects",
				"policy":  "modify_production",
			})
			return
		}
	}

	if req.Name != nil {
		proj.Name = *req.Name
	}
	if req.Description != nil {
		proj.Description = *req.Description
	}
	if req.Environment != nil {
		proj.Environment = *req.Environment
	}
	if req.Status != nil {
		proj.Status = *req.Status
	}
	if req.Tags != nil {
		proj.Tags = req.Tags
	}

	proj.UpdatedAt = time.Now()
	h.store.UpdateProject(proj)

	c.JSON(http.StatusOK, gin.H{
		"project":     proj,
		"permissions": h.evaluateABACPolicies(userCtx, proj),
	})
}

// Delete deletes a project
// DELETE /api/v1/projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)
	projID := c.Param("id")

	proj, err := h.store.GetProject(projID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	permissions := h.evaluateABACPolicies(userCtx, proj)

	// ABAC Policy: Check delete permission
	if !permissions["can_delete"] {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "policy_violation",
			"message": "You don't have permission to delete this project",
			"reason":  h.getDeleteDenialReason(userCtx, proj),
		})
		return
	}

	h.store.DeleteProject(projID)
	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}

// Deploy triggers a deployment for the project
// POST /api/v1/projects/:id/deploy
func (h *ProjectHandler) Deploy(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)
	projID := c.Param("id")

	proj, err := h.store.GetProject(projID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	permissions := h.evaluateABACPolicies(userCtx, proj)

	// ABAC Policy: Check deploy permission
	if !permissions["can_deploy"] {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "policy_violation",
			"message": "You don't have permission to deploy this project",
			"reason":  h.getDeployDenialReason(userCtx, proj),
			"policies": []string{
				"Only admins can deploy to production",
				"Project must be in 'active' status",
			},
		})
		return
	}

	// Simulate deployment
	c.JSON(http.StatusOK, gin.H{
		"message":     "Deployment initiated",
		"project_id":  proj.ID,
		"environment": proj.Environment,
		"deployed_by": userCtx.UserID,
		"deployed_at": time.Now(),
	})
}

// ABAC Policy Evaluation

func (h *ProjectHandler) evaluateABACPolicies(userCtx *store.UserContext, proj *store.Project) map[string]bool {
	isAdmin := h.isAdmin(userCtx)
	isOwner := proj.OwnerID == userCtx.UserID
	isProduction := proj.Environment == "production"
	isArchived := proj.Status == "archived"
	isPaused := proj.Status == "paused"

	// Base permissions for workspace members
	canRead := true
	canWrite := true
	canDelete := false
	canDeploy := true

	// Policy: Archived projects are read-only
	if isArchived {
		canWrite = false
		canDelete = false
		canDeploy = false
	}

	// Policy: Paused projects can't be deployed
	if isPaused {
		canDeploy = false
	}

	// Policy: Production projects have stricter rules
	if isProduction {
		// Only admins can modify production
		if !isAdmin {
			canWrite = false
			canDeploy = false
		}
		// Only admins can delete production projects
		canDelete = isAdmin
	} else {
		// Non-production: owners and admins can delete
		canDelete = isOwner || isAdmin
	}

	// Policy: Admins can always deploy to non-archived projects
	if isAdmin && !isArchived {
		canDeploy = true
	}

	// Platform admins override all
	if userCtx.IsPlatformAdmin {
		canRead = true
		canWrite = true
		canDelete = true
		canDeploy = !isArchived
	}

	return map[string]bool{
		"can_read":   canRead,
		"can_write":  canWrite,
		"can_delete": canDelete,
		"can_deploy": canDeploy,
	}
}

func (h *ProjectHandler) isAdmin(userCtx *store.UserContext) bool {
	// In real app, check workspace role from OpenFGA or database
	// For demo, use platform admin flag
	return userCtx.IsPlatformAdmin
}

func (h *ProjectHandler) getWriteDenialReason(userCtx *store.UserContext, proj *store.Project) string {
	if proj.Status == "archived" {
		return "Archived projects cannot be modified"
	}
	if proj.Environment == "production" && !h.isAdmin(userCtx) {
		return "Only administrators can modify production projects"
	}
	return "Insufficient permissions"
}

func (h *ProjectHandler) getDeleteDenialReason(userCtx *store.UserContext, proj *store.Project) string {
	if proj.Status == "archived" {
		return "Archived projects cannot be deleted"
	}
	if proj.Environment == "production" && !h.isAdmin(userCtx) {
		return "Only administrators can delete production projects"
	}
	if proj.OwnerID != userCtx.UserID && !h.isAdmin(userCtx) {
		return "Only the owner or administrators can delete projects"
	}
	return "Insufficient permissions"
}

func (h *ProjectHandler) getDeployDenialReason(userCtx *store.UserContext, proj *store.Project) string {
	if proj.Status == "archived" {
		return "Archived projects cannot be deployed"
	}
	if proj.Status == "paused" {
		return "Paused projects cannot be deployed"
	}
	if proj.Environment == "production" && !h.isAdmin(userCtx) {
		return "Only administrators can deploy to production"
	}
	return "Insufficient permissions"
}

func (h *ProjectHandler) getActivePolicies() []map[string]string {
	return []map[string]string{
		{
			"name":        "production_admin_only",
			"description": "Only administrators can modify production projects",
		},
		{
			"name":        "archived_read_only",
			"description": "Archived projects are read-only",
		},
		{
			"name":        "paused_no_deploy",
			"description": "Paused projects cannot be deployed",
		},
		{
			"name":        "owner_can_delete",
			"description": "Project owners can delete non-production projects",
		},
	}
}
