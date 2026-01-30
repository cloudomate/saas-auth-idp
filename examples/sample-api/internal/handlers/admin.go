package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/sample-api/internal/store"
)

// AdminHandler handles platform admin operations
type AdminHandler struct {
	store *store.MemoryStore
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(s *store.MemoryStore) *AdminHandler {
	return &AdminHandler{store: s}
}

// GetStats returns platform-wide statistics
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats := h.store.GetPlatformStats()
	c.JSON(http.StatusOK, stats)
}

// ListUsers returns all users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	users := h.store.ListUsers()
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUser returns a specific user
func (h *AdminHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.store.GetUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateUserAdmin updates a user's platform admin status
func (h *AdminHandler) UpdateUserAdmin(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		IsPlatformAdmin bool `json:"is_platform_admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.store.SetPlatformAdmin(id, req.IsPlatformAdmin); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user, _ := h.store.GetUser(id)
	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// Prevent deleting yourself
	currentUserID := c.GetString("user_id")
	if id == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}

	if err := h.store.DeleteUser(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// ListTenants returns all tenants
func (h *AdminHandler) ListTenants(c *gin.Context) {
	tenants := h.store.ListTenants()
	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

// GetTenant returns a specific tenant
func (h *AdminHandler) GetTenant(c *gin.Context) {
	id := c.Param("id")
	tenant, err := h.store.GetTenant(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	// Get workspaces for this tenant
	workspaces := h.store.ListWorkspacesByTenant(id)

	c.JSON(http.StatusOK, gin.H{
		"tenant":     tenant,
		"workspaces": workspaces,
	})
}

// DeleteTenant deletes a tenant
func (h *AdminHandler) DeleteTenant(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.DeleteTenant(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tenant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tenant deleted"})
}

// ListWorkspaces returns all workspaces
func (h *AdminHandler) ListWorkspaces(c *gin.Context) {
	workspaces := h.store.ListWorkspaces()
	c.JSON(http.StatusOK, gin.H{"workspaces": workspaces})
}

// DeleteWorkspace deletes a workspace
func (h *AdminHandler) DeleteWorkspace(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.DeleteWorkspace(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "workspace deleted"})
}

// ListAllDocuments returns all documents across all workspaces
func (h *AdminHandler) ListAllDocuments(c *gin.Context) {
	docs := h.store.GetAllDocuments()
	c.JSON(http.StatusOK, gin.H{"documents": docs})
}

// ListAllProjects returns all projects across all workspaces
func (h *AdminHandler) ListAllProjects(c *gin.Context) {
	projects := h.store.GetAllProjects()
	c.JSON(http.StatusOK, gin.H{"projects": projects})
}
