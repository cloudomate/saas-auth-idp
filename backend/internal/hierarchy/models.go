package hierarchy

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResourceContainer represents a generic container at any level of the hierarchy
// This unified model replaces separate Tenant/Workspace models
type ResourceContainer struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Level       string     `gorm:"index;not null" json:"level"`                         // e.g., "tenant", "workspace", "project"
	Slug        string     `gorm:"not null" json:"slug"`                                // URL-friendly identifier
	DisplayName string     `gorm:"not null" json:"display_name"`                        // Human-readable name
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id,omitempty"`          // Parent container (nil for root)
	RootID      uuid.UUID  `gorm:"type:uuid;index;not null" json:"root_id"`             // Root tenant ID (for fast queries)
	Path        string     `gorm:"index" json:"path"`                                   // Materialized path: /root-id/parent-id/id
	Depth       int        `gorm:"not null" json:"depth"`                               // Depth in hierarchy (0 = root)
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	Metadata    string     `gorm:"type:jsonb" json:"metadata,omitempty"`                // Flexible JSON metadata
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	Parent   *ResourceContainer  `gorm:"foreignKey:ParentID" json:"-"`
	Children []ResourceContainer `gorm:"foreignKey:ParentID" json:"-"`
}

// TableName returns the table name for GORM
func (ResourceContainer) TableName() string {
	return "resource_containers"
}

// ContainerMembership links users to containers with roles
type ContainerMembership struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	ContainerID uuid.UUID `gorm:"type:uuid;index;not null" json:"container_id"`
	Role        string    `gorm:"not null;default:'member'" json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User      User              `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Container ResourceContainer `gorm:"foreignKey:ContainerID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName returns the table name for GORM
func (ContainerMembership) TableName() string {
	return "container_memberships"
}

// User model (simplified, keeping auth fields)
type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email           string     `gorm:"uniqueIndex;not null" json:"email"`
	Name            string     `json:"name"`
	Picture         string     `json:"picture,omitempty"`
	IsPlatformAdmin bool       `gorm:"default:false" json:"is_platform_admin"`
	AuthProvider    string     `gorm:"type:text" json:"auth_provider"`
	EmailVerified   bool       `gorm:"default:false" json:"email_verified"`
	PasswordHash    string     `gorm:"type:text" json:"-"`
	VerifyToken     string     `gorm:"type:text" json:"-"`
	VerifyExpiry    *time.Time `json:"-"`
	ResetToken      string     `gorm:"type:text" json:"-"`
	ResetExpiry     *time.Time `json:"-"`

	// Root container admin (tenant admin)
	AdminOfRootID    *uuid.UUID `gorm:"type:uuid;index" json:"admin_of_root_id,omitempty"`
	IsRootAdmin      bool       `gorm:"default:false" json:"is_root_admin"`
	SelectedPlanTier string     `gorm:"type:varchar(20)" json:"selected_plan,omitempty"`

	LastLogin time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Memberships []ContainerMembership `gorm:"foreignKey:UserID" json:"-"`
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

// Repository provides database operations for the hierarchy
type Repository struct {
	db     *gorm.DB
	config *Config
}

// NewRepository creates a new hierarchy repository
func NewRepository(db *gorm.DB, config *Config) *Repository {
	return &Repository{db: db, config: config}
}

// CreateContainer creates a new container at any level
func (r *Repository) CreateContainer(level, slug, displayName string, parentID *uuid.UUID) (*ResourceContainer, error) {
	container := &ResourceContainer{
		Level:       level,
		Slug:        slug,
		DisplayName: displayName,
		ParentID:    parentID,
		IsActive:    true,
	}

	// Calculate depth and path
	if parentID == nil {
		// Root container
		container.Depth = 0
		container.RootID = container.ID // Will be set after create
	} else {
		// Child container
		var parent ResourceContainer
		if err := r.db.First(&parent, "id = ?", parentID).Error; err != nil {
			return nil, err
		}
		container.Depth = parent.Depth + 1
		container.RootID = parent.RootID
		container.Path = parent.Path
	}

	if err := r.db.Create(container).Error; err != nil {
		return nil, err
	}

	// Update path and root_id for new container
	if parentID == nil {
		container.RootID = container.ID
		container.Path = "/" + container.ID.String()
	} else {
		container.Path = container.Path + "/" + container.ID.String()
	}
	r.db.Save(container)

	return container, nil
}

// GetContainer retrieves a container by ID
func (r *Repository) GetContainer(id uuid.UUID) (*ResourceContainer, error) {
	var container ResourceContainer
	if err := r.db.First(&container, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &container, nil
}

// GetContainerBySlug retrieves a container by slug within a parent
func (r *Repository) GetContainerBySlug(level, slug string, parentID *uuid.UUID) (*ResourceContainer, error) {
	var container ResourceContainer
	query := r.db.Where("level = ? AND slug = ?", level, slug)
	if parentID != nil {
		query = query.Where("parent_id = ?", parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	if err := query.First(&container).Error; err != nil {
		return nil, err
	}
	return &container, nil
}

// ListChildren lists all direct children of a container
func (r *Repository) ListChildren(parentID uuid.UUID, level string) ([]ResourceContainer, error) {
	var containers []ResourceContainer
	query := r.db.Where("parent_id = ?", parentID)
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if err := query.Order("created_at ASC").Find(&containers).Error; err != nil {
		return nil, err
	}
	return containers, nil
}

// ListByRoot lists all containers under a root tenant
func (r *Repository) ListByRoot(rootID uuid.UUID, level string) ([]ResourceContainer, error) {
	var containers []ResourceContainer
	query := r.db.Where("root_id = ?", rootID)
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if err := query.Order("depth ASC, created_at ASC").Find(&containers).Error; err != nil {
		return nil, err
	}
	return containers, nil
}

// GetAncestors returns all ancestors of a container (from parent to root)
func (r *Repository) GetAncestors(containerID uuid.UUID) ([]ResourceContainer, error) {
	var container ResourceContainer
	if err := r.db.First(&container, "id = ?", containerID).Error; err != nil {
		return nil, err
	}

	var ancestors []ResourceContainer
	currentID := container.ParentID
	for currentID != nil {
		var parent ResourceContainer
		if err := r.db.First(&parent, "id = ?", currentID).Error; err != nil {
			break
		}
		ancestors = append(ancestors, parent)
		currentID = parent.ParentID
	}

	return ancestors, nil
}

// AddMember adds a user to a container with a role
func (r *Repository) AddMember(userID, containerID uuid.UUID, role string) error {
	membership := &ContainerMembership{
		UserID:      userID,
		ContainerID: containerID,
		Role:        role,
	}
	return r.db.Create(membership).Error
}

// GetMembership gets a user's membership in a container
func (r *Repository) GetMembership(userID, containerID uuid.UUID) (*ContainerMembership, error) {
	var membership ContainerMembership
	if err := r.db.Where("user_id = ? AND container_id = ?", userID, containerID).First(&membership).Error; err != nil {
		return nil, err
	}
	return &membership, nil
}

// ListMembers lists all members of a container
func (r *Repository) ListMembers(containerID uuid.UUID) ([]ContainerMembership, error) {
	var memberships []ContainerMembership
	if err := r.db.Preload("User").Where("container_id = ?", containerID).Find(&memberships).Error; err != nil {
		return nil, err
	}
	return memberships, nil
}

// GetUserContainers lists all containers a user has access to at a given level
func (r *Repository) GetUserContainers(userID uuid.UUID, level string) ([]ResourceContainer, error) {
	var containers []ResourceContainer
	query := `
		SELECT DISTINCT rc.* FROM resource_containers rc
		JOIN container_memberships cm ON rc.id = cm.container_id
		WHERE cm.user_id = ? AND rc.level = ?
		ORDER BY rc.created_at ASC
	`
	if err := r.db.Raw(query, userID, level).Scan(&containers).Error; err != nil {
		return nil, err
	}
	return containers, nil
}

// AutoMigrate runs database migrations for hierarchy models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&ResourceContainer{},
		&ContainerMembership{},
	)
}
