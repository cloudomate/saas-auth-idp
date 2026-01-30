package store

import "time"

// User represents a platform user
type User struct {
	ID              string    `json:"id"`
	Email           string    `json:"email"`
	Name            string    `json:"name"`
	Picture         string    `json:"picture,omitempty"`
	IsPlatformAdmin bool      `json:"is_platform_admin"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Tenant represents an organization/company
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Plan      string    `json:"plan"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Workspace represents a workspace within a tenant
type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	TenantID  string    `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PlatformStats represents platform-wide statistics
type PlatformStats struct {
	TotalUsers      int `json:"total_users"`
	TotalTenants    int `json:"total_tenants"`
	TotalWorkspaces int `json:"total_workspaces"`
	TotalDocuments  int `json:"total_documents"`
	TotalProjects   int `json:"total_projects"`
	AdminCount      int `json:"admin_count"`
}

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

// SocialProvider represents a social login provider configuration
type SocialProvider struct {
	ID           string    `json:"id"`
	Provider     string    `json:"provider"` // google, github, microsoft, etc.
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"-"` // Never expose in JSON
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SSOConfig represents an enterprise OIDC/SAML SSO configuration
type SSOConfig struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id,omitempty"` // Empty for platform-level SSO
	Name            string    `json:"name"`
	Type            string    `json:"type"` // oidc, saml
	Enabled         bool      `json:"enabled"`
	IssuerURL       string    `json:"issuer_url,omitempty"`
	ClientID        string    `json:"client_id,omitempty"`
	ClientSecret    string    `json:"-"` // Never expose in JSON
	AuthorizationURL string   `json:"authorization_url,omitempty"`
	TokenURL        string    `json:"token_url,omitempty"`
	UserInfoURL     string    `json:"userinfo_url,omitempty"`
	Scopes          []string  `json:"scopes,omitempty"`
	// SAML specific
	MetadataURL     string    `json:"metadata_url,omitempty"`
	EntityID        string    `json:"entity_id,omitempty"`
	ACSURL          string    `json:"acs_url,omitempty"`
	Certificate     string    `json:"-"` // Never expose in JSON
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AdminInviteToken represents an invite token for platform admin registration
type AdminInviteToken struct {
	ID        string    `json:"id"`
	Token     string    `json:"token"`
	Email     string    `json:"email,omitempty"` // Optional: restrict to specific email
	Used      bool      `json:"used"`
	UsedBy    string    `json:"used_by,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// PlatformSettings represents global platform configuration
type PlatformSettings struct {
	ID                    string    `json:"id"`
	PlatformName          string    `json:"platform_name"`
	AllowUserRegistration bool      `json:"allow_user_registration"`
	RequireEmailVerify    bool      `json:"require_email_verification"`
	DefaultUserRole       string    `json:"default_user_role"`
	SessionTimeout        int       `json:"session_timeout"` // in minutes
	UpdatedAt             time.Time `json:"updated_at"`
}
