package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/sample-api/internal/authz"
	"github.com/yourusername/sample-api/internal/middleware"
	"github.com/yourusername/sample-api/internal/store"
)

// DocumentHandler handles document operations
// This demonstrates ReBAC (Relationship-Based Access Control)
//
// ReBAC Pattern:
// - Access is determined by relationships between users and documents
// - Relationships: owner, editor, viewer
// - Permissions inherit from relationships:
//   - owner: can_read, can_write, can_delete, can_share
//   - editor: can_read, can_write
//   - viewer: can_read
type DocumentHandler struct {
	store *store.MemoryStore
	fga   *authz.OpenFGAClient
}

func NewDocumentHandler(s *store.MemoryStore, fga *authz.OpenFGAClient) *DocumentHandler {
	return &DocumentHandler{store: s, fga: fga}
}

// List returns documents the user can access
// GET /api/v1/documents
func (h *DocumentHandler) List(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)

	// Get documents user can see based on visibility and sharing
	docs := h.store.ListDocumentsForUser(userCtx.WorkspaceID, userCtx.UserID)

	// Enrich with user's permission level
	type DocWithPermissions struct {
		*store.Document
		Permissions map[string]bool `json:"permissions"`
	}

	result := make([]DocWithPermissions, 0, len(docs))
	for _, doc := range docs {
		permissions := h.getUserPermissions(userCtx, doc)
		result = append(result, DocWithPermissions{
			Document:    doc,
			Permissions: permissions,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": result,
		"total":     len(result),
	})
}

// Create creates a new document
// POST /api/v1/documents
func (h *DocumentHandler) Create(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)

	var req struct {
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content"`
		Visibility string `json:"visibility"` // public, workspace, private
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	if req.Visibility == "" {
		req.Visibility = "workspace"
	}

	doc := &store.Document{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Content:     req.Content,
		WorkspaceID: userCtx.WorkspaceID,
		OwnerID:     userCtx.UserID,
		Visibility:  req.Visibility,
		Status:      "draft",
	}

	if err := h.store.CreateDocument(doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create document"})
		return
	}

	// Create OpenFGA relationship for owner
	if h.fga != nil {
		// document:doc-id#owner@user:user-id
		h.fga.WriteTuple(
			fmt.Sprintf("user:%s", userCtx.UserID),
			"owner",
			fmt.Sprintf("document:%s", doc.ID),
		)
		// document:doc-id#container@container:workspace-id
		h.fga.WriteTuple(
			fmt.Sprintf("container:%s", userCtx.WorkspaceID),
			"container",
			fmt.Sprintf("document:%s", doc.ID),
		)
	}

	c.JSON(http.StatusCreated, gin.H{
		"document": doc,
		"permissions": map[string]bool{
			"can_read":   true,
			"can_write":  true,
			"can_delete": true,
			"can_share":  true,
		},
	})
}

// Get returns a specific document
// GET /api/v1/documents/:id
func (h *DocumentHandler) Get(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)
	docID := c.Param("id")

	doc, err := h.store.GetDocument(docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	// Check access using ReBAC
	if !h.canRead(userCtx, doc) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": "You don't have permission to view this document",
		})
		return
	}

	permissions := h.getUserPermissions(userCtx, doc)

	c.JSON(http.StatusOK, gin.H{
		"document":    doc,
		"permissions": permissions,
		"shares":      h.store.GetDocumentShares(docID),
	})
}

// Update updates a document
// PUT /api/v1/documents/:id
func (h *DocumentHandler) Update(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)
	docID := c.Param("id")

	doc, err := h.store.GetDocument(docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	// Check write permission using ReBAC
	if !h.canWrite(userCtx, doc) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": "You don't have permission to edit this document",
		})
		return
	}

	var req struct {
		Title      *string `json:"title"`
		Content    *string `json:"content"`
		Visibility *string `json:"visibility"`
		Status     *string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Title != nil {
		doc.Title = *req.Title
	}
	if req.Content != nil {
		doc.Content = *req.Content
	}
	if req.Visibility != nil {
		// Only owner can change visibility
		if doc.OwnerID != userCtx.UserID {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "access_denied",
				"message": "Only the owner can change document visibility",
			})
			return
		}
		doc.Visibility = *req.Visibility
	}
	if req.Status != nil {
		doc.Status = *req.Status
	}

	doc.UpdatedAt = time.Now()
	h.store.UpdateDocument(doc)

	c.JSON(http.StatusOK, gin.H{"document": doc})
}

// Delete deletes a document
// DELETE /api/v1/documents/:id
func (h *DocumentHandler) Delete(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)
	docID := c.Param("id")

	doc, err := h.store.GetDocument(docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	// Only owner can delete
	if !h.canDelete(userCtx, doc) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": "Only the owner can delete this document",
		})
		return
	}

	h.store.DeleteDocument(docID)

	// Remove OpenFGA relationships
	if h.fga != nil {
		h.fga.DeleteTuple(
			fmt.Sprintf("user:%s", doc.OwnerID),
			"owner",
			fmt.Sprintf("document:%s", docID),
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "document deleted"})
}

// Share shares a document with another user
// POST /api/v1/documents/:id/share
func (h *DocumentHandler) Share(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)
	docID := c.Param("id")

	doc, err := h.store.GetDocument(docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	// Only owner can share
	if !h.canShare(userCtx, doc) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": "Only the owner can share this document",
		})
		return
	}

	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"` // editor, viewer
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Role != "editor" && req.Role != "viewer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be 'editor' or 'viewer'"})
		return
	}

	share := store.DocumentShare{
		DocumentID: docID,
		UserID:     req.UserID,
		Role:       req.Role,
	}

	if err := h.store.AddDocumentShare(share); err != nil {
		if err == store.ErrAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "user already has access"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to share document"})
		return
	}

	// Create OpenFGA relationship
	if h.fga != nil {
		h.fga.WriteTuple(
			fmt.Sprintf("user:%s", req.UserID),
			req.Role,
			fmt.Sprintf("document:%s", docID),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Document shared with user as %s", req.Role),
		"share":   share,
	})
}

// GetPermissions returns the current user's permissions on a document
// GET /api/v1/documents/:id/permissions
func (h *DocumentHandler) GetPermissions(c *gin.Context) {
	userCtx := middleware.GetUserContext(c)
	docID := c.Param("id")

	doc, err := h.store.GetDocument(docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	permissions := h.getUserPermissions(userCtx, doc)
	role := h.store.GetUserDocumentRole(docID, userCtx.UserID)

	c.JSON(http.StatusOK, gin.H{
		"user_id":     userCtx.UserID,
		"document_id": docID,
		"role":        role,
		"permissions": permissions,
	})
}

// Permission check helpers - ReBAC logic

func (h *DocumentHandler) getUserPermissions(userCtx *store.UserContext, doc *store.Document) map[string]bool {
	// Check via OpenFGA if available
	if h.fga != nil {
		relations, err := h.fga.ListRelations(
			fmt.Sprintf("user:%s", userCtx.UserID),
			fmt.Sprintf("document:%s", doc.ID),
			[]string{"can_read", "can_write", "can_manage"},
		)
		if err == nil {
			return map[string]bool{
				"can_read":   relations["can_read"],
				"can_write":  relations["can_write"],
				"can_delete": relations["can_manage"],
				"can_share":  relations["can_manage"],
			}
		}
	}

	// Fallback to local logic
	return map[string]bool{
		"can_read":   h.canRead(userCtx, doc),
		"can_write":  h.canWrite(userCtx, doc),
		"can_delete": h.canDelete(userCtx, doc),
		"can_share":  h.canShare(userCtx, doc),
	}
}

func (h *DocumentHandler) canRead(userCtx *store.UserContext, doc *store.Document) bool {
	// Platform admin can read everything
	if userCtx.IsPlatformAdmin {
		return true
	}

	// Public documents are readable by anyone in workspace
	if doc.Visibility == "public" && doc.WorkspaceID == userCtx.WorkspaceID {
		return true
	}

	// Workspace-visible documents are readable by workspace members
	if doc.Visibility == "workspace" && doc.WorkspaceID == userCtx.WorkspaceID {
		return true
	}

	// Owner can always read
	if doc.OwnerID == userCtx.UserID {
		return true
	}

	// Check explicit share
	role := h.store.GetUserDocumentRole(doc.ID, userCtx.UserID)
	return role != ""
}

func (h *DocumentHandler) canWrite(userCtx *store.UserContext, doc *store.Document) bool {
	if userCtx.IsPlatformAdmin {
		return true
	}

	if doc.OwnerID == userCtx.UserID {
		return true
	}

	role := h.store.GetUserDocumentRole(doc.ID, userCtx.UserID)
	return role == "owner" || role == "editor"
}

func (h *DocumentHandler) canDelete(userCtx *store.UserContext, doc *store.Document) bool {
	if userCtx.IsPlatformAdmin {
		return true
	}

	return doc.OwnerID == userCtx.UserID
}

func (h *DocumentHandler) canShare(userCtx *store.UserContext, doc *store.Document) bool {
	if userCtx.IsPlatformAdmin {
		return true
	}

	return doc.OwnerID == userCtx.UserID
}
