package auth

// Identity represents the authenticated user/service
type Identity struct {
	UserID          string
	Email           string
	TenantID        string
	WorkspaceID     string
	Role            string
	IsPlatformAdmin bool
	KeyID           string // For API keys
}
