package store

import "time"

// Document represents a document resource (for ReBAC demo)
// ReBAC: Access is based on relationships (owner, editor, viewer)
type Document struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	WorkspaceID string    `json:"workspace_id"`
	OwnerID     string    `json:"owner_id"`
	Visibility  string    `json:"visibility"` // public, workspace, private
	Status      string    `json:"status"`     // draft, published, archived
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DocumentShare represents a sharing relationship
type DocumentShare struct {
	DocumentID string `json:"document_id"`
	UserID     string `json:"user_id"`
	Role       string `json:"role"` // owner, editor, viewer
}

// Project represents a project resource (for ABAC demo)
// ABAC: Access is based on attributes (environment, status, tags)
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	WorkspaceID string    `json:"workspace_id"`
	OwnerID     string    `json:"owner_id"`
	Environment string    `json:"environment"` // production, staging, development
	Status      string    `json:"status"`      // active, paused, archived
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserContext represents the authenticated user context
type UserContext struct {
	UserID          string   `json:"user_id"`
	TenantID        string   `json:"tenant_id"`
	WorkspaceID     string   `json:"workspace_id"`
	IsPlatformAdmin bool     `json:"is_platform_admin"`
	Roles           []string `json:"roles"` // workspace roles
}

// PermissionCheck represents a permission check result
type PermissionCheck struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}
