package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// User Model
// ============================================================================

// User represents a user in the system (unified: platform signups and tenant users)
type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email           string     `gorm:"uniqueIndex;not null" json:"email"`
	Name            string     `json:"name"`
	Picture         string     `json:"picture,omitempty"`
	IsPlatformAdmin bool       `gorm:"default:false" json:"is_platform_admin"`

	// Auth fields
	AuthProvider  string     `gorm:"type:text" json:"auth_provider"` // "google", "github", "local"
	EmailVerified bool       `gorm:"default:false" json:"email_verified"`
	PasswordHash  string     `gorm:"type:text" json:"-"`             // For local auth
	VerifyToken   string     `gorm:"type:text" json:"-"`
	VerifyExpiry  *time.Time `json:"-"`
	ResetToken    string     `gorm:"type:text" json:"-"`
	ResetExpiry   *time.Time `json:"-"`

	// Tenant admin fields
	IsTenantAdmin       bool       `gorm:"default:false" json:"is_tenant_admin"`
	AdminOfTenantID     *uuid.UUID `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	SelectedPlanTier    PlanTier   `gorm:"type:varchar(20)" json:"selected_plan,omitempty"`

	// Timestamps
	LastLogin time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	AdminOfTenant *Tenant      `gorm:"foreignKey:AdminOfTenantID" json:"-"`
	Memberships   []Membership `gorm:"foreignKey:UserID" json:"-"`
}

// IsVerifyTokenExpired checks if the verification token has expired
func (u *User) IsVerifyTokenExpired() bool {
	if u.VerifyExpiry == nil {
		return true
	}
	return time.Now().After(*u.VerifyExpiry)
}

// IsResetTokenExpired checks if the reset token has expired
func (u *User) IsResetTokenExpired() bool {
	if u.ResetExpiry == nil {
		return true
	}
	return time.Now().After(*u.ResetExpiry)
}

// ============================================================================
// Tenant Model
// ============================================================================

// Tenant represents an organization/company
type Tenant struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	DisplayName string    `gorm:"not null" json:"display_name"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	Metadata    string    `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Admin user
	AdminUserID *uuid.UUID `gorm:"type:uuid;index" json:"admin_user_id,omitempty"`

	// Subscription
	SubscriptionID *uuid.UUID `gorm:"type:uuid;index" json:"-"`

	// SSO Configuration
	SSOConfigured bool `gorm:"default:false" json:"sso_configured"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Workspaces   []Workspace   `gorm:"foreignKey:TenantID" json:"-"`
	Subscription *Subscription `gorm:"foreignKey:TenantID" json:"-"`
}

// ============================================================================
// Workspace Model
// ============================================================================

// Workspace represents a project/team within a tenant
type Workspace struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID `gorm:"type:uuid;index;not null" json:"tenant_id"`
	Slug        string    `gorm:"not null" json:"slug"`
	DisplayName string    `json:"display_name"`
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Tenant      Tenant       `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE" json:"-"`
	Memberships []Membership `gorm:"foreignKey:WorkspaceID" json:"-"`
}

// ============================================================================
// Membership Model
// ============================================================================

// Membership links users to workspaces with roles
type Membership struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;index;not null" json:"workspace_id"`
	Role        string    `gorm:"not null;default:'member'" json:"role"` // admin, member, viewer
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Workspace Workspace `gorm:"foreignKey:WorkspaceID;constraint:OnDelete:CASCADE" json:"-"`
}

// ============================================================================
// Subscription & Plan Models
// ============================================================================

// PlanTier represents the subscription tier
type PlanTier string

const (
	PlanTierBasic      PlanTier = "basic"
	PlanTierAdvanced   PlanTier = "advanced"
	PlanTierEnterprise PlanTier = "enterprise"
)

// Plan represents a subscription plan
type Plan struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Tier              PlanTier  `gorm:"type:varchar(20);uniqueIndex;not null" json:"tier"`
	Name              string    `gorm:"not null" json:"name"`
	Description       string    `json:"description"`
	MaxWorkspaces     int       `gorm:"default:-1" json:"max_workspaces"`   // -1 = unlimited
	MaxUsersPerTenant int       `gorm:"default:-1" json:"max_users"`        // -1 = unlimited
	MonthlyPriceCents int       `gorm:"default:0" json:"monthly_price"`
	AnnualPriceCents  int       `gorm:"default:0" json:"annual_price"`
	AllowsOnPrem      bool      `gorm:"default:false" json:"allows_on_prem"`
	Features          string    `gorm:"type:jsonb" json:"features"`         // JSON array of feature strings
	IsActive          bool      `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Subscription links a tenant to a plan
type Subscription struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID             uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"tenant_id"`
	PlanID               uuid.UUID `gorm:"type:uuid;not null" json:"plan_id"`
	Status               string    `gorm:"default:'active'" json:"status"` // active, cancelled, past_due, trialing
	CurrentPeriodStart   time.Time `json:"current_period_start"`
	CurrentPeriodEnd     time.Time `json:"current_period_end"`
	StripeCustomerID     string    `json:"-"`
	StripeSubscriptionID string    `json:"-"`
	CancelledAt          *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Relationships
	Tenant Tenant `gorm:"foreignKey:TenantID;constraint:OnDelete:CASCADE" json:"-"`
	Plan   Plan   `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

// ============================================================================
// OAuth State Model (for CSRF protection)
// ============================================================================

// OAuthState stores OAuth state for CSRF protection
type OAuthState struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	State     string    `gorm:"uniqueIndex;not null"`
	Provider  string    `gorm:"not null"` // google, github
	Plan      string    // Optional: plan tier selected during signup
	Flow      string    // signup, login
	ExpiresAt time.Time
	CreatedAt time.Time
}

// ============================================================================
// Database Migration
// ============================================================================

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Tenant{},
		&Workspace{},
		&Membership{},
		&Plan{},
		&Subscription{},
		&OAuthState{},
	)
}

// SeedPlans creates default subscription plans
func SeedPlans(db *gorm.DB) error {
	plans := []Plan{
		{
			Tier:              PlanTierBasic,
			Name:              "Basic",
			Description:       "For individual users",
			MaxWorkspaces:     1,
			MaxUsersPerTenant: 1,
			MonthlyPriceCents: 0,
			AnnualPriceCents:  0,
			AllowsOnPrem:      false,
			Features:          `["Core features", "Community support"]`,
			IsActive:          true,
		},
		{
			Tier:              PlanTierAdvanced,
			Name:              "Advanced",
			Description:       "For small teams",
			MaxWorkspaces:     5,
			MaxUsersPerTenant: 10,
			MonthlyPriceCents: 4900,
			AnnualPriceCents:  49000,
			AllowsOnPrem:      false,
			Features:          `["Everything in Basic", "SSO configuration", "Priority support", "API access"]`,
			IsActive:          true,
		},
		{
			Tier:              PlanTierEnterprise,
			Name:              "Enterprise",
			Description:       "For large organizations",
			MaxWorkspaces:     -1,
			MaxUsersPerTenant: -1,
			MonthlyPriceCents: 0, // Contact sales
			AnnualPriceCents:  0,
			AllowsOnPrem:      true,
			Features:          `["Everything in Advanced", "Unlimited workspaces", "Unlimited users", "On-premises deployment", "Dedicated support", "Custom integrations"]`,
			IsActive:          true,
		},
	}

	for _, plan := range plans {
		var existing Plan
		if err := db.Where("tier = ?", plan.Tier).First(&existing).Error; err != nil {
			if err := db.Create(&plan).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
