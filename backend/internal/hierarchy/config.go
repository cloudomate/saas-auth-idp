package hierarchy

import (
	"encoding/json"
	"os"
)

// Level represents a single level in the hierarchy
type Level struct {
	Name        string   `json:"name"`         // Internal name (e.g., "workspace", "project")
	DisplayName string   `json:"display_name"` // UI display name (e.g., "Workspace", "Project")
	Plural      string   `json:"plural"`       // Plural form (e.g., "workspaces", "projects")
	URLPath     string   `json:"url_path"`     // API path segment (e.g., "workspaces", "projects")
	Roles       []string `json:"roles"`        // Available roles at this level
	IsRoot      bool     `json:"is_root"`      // Is this the root level (tenant)?
}

// Config defines the complete hierarchy configuration
type Config struct {
	// Levels defines the hierarchy from root to leaf
	// Example: [tenant, team, project] or [organization, workspace]
	Levels []Level `json:"levels"`

	// RootLevel is the top-level container (typically "tenant" or "organization")
	RootLevel string `json:"root_level"`

	// LeafLevel is the level where resources are scoped (typically "workspace" or "project")
	LeafLevel string `json:"leaf_level"`
}

// DefaultConfig returns the default 2-level hierarchy: Tenant → Workspace
func DefaultConfig() *Config {
	return &Config{
		RootLevel: "tenant",
		LeafLevel: "workspace",
		Levels: []Level{
			{
				Name:        "tenant",
				DisplayName: "Organization",
				Plural:      "organizations",
				URLPath:     "tenant",
				Roles:       []string{"admin", "member"},
				IsRoot:      true,
			},
			{
				Name:        "workspace",
				DisplayName: "Workspace",
				Plural:      "workspaces",
				URLPath:     "workspaces",
				Roles:       []string{"admin", "member", "viewer"},
				IsRoot:      false,
			},
		},
	}
}

// MLPlatformConfig returns a 3-level hierarchy for ML platforms: Tenant → Team → Project
func MLPlatformConfig() *Config {
	return &Config{
		RootLevel: "tenant",
		LeafLevel: "project",
		Levels: []Level{
			{
				Name:        "tenant",
				DisplayName: "Organization",
				Plural:      "organizations",
				URLPath:     "tenant",
				Roles:       []string{"admin", "member"},
				IsRoot:      true,
			},
			{
				Name:        "team",
				DisplayName: "Team",
				Plural:      "teams",
				URLPath:     "teams",
				Roles:       []string{"admin", "member"},
				IsRoot:      false,
			},
			{
				Name:        "project",
				DisplayName: "Project",
				Plural:      "projects",
				URLPath:     "projects",
				Roles:       []string{"admin", "member", "viewer"},
				IsRoot:      false,
			},
		},
	}
}

// DevOpsConfig returns a hierarchy for DevOps platforms: Tenant → Environment → Service
func DevOpsConfig() *Config {
	return &Config{
		RootLevel: "tenant",
		LeafLevel: "service",
		Levels: []Level{
			{
				Name:        "tenant",
				DisplayName: "Organization",
				Plural:      "organizations",
				URLPath:     "tenant",
				Roles:       []string{"admin", "member"},
				IsRoot:      true,
			},
			{
				Name:        "environment",
				DisplayName: "Environment",
				Plural:      "environments",
				URLPath:     "environments",
				Roles:       []string{"admin", "operator"},
				IsRoot:      false,
			},
			{
				Name:        "service",
				DisplayName: "Service",
				Plural:      "services",
				URLPath:     "services",
				Roles:       []string{"admin", "developer", "viewer"},
				IsRoot:      false,
			},
		},
	}
}

// LoadFromFile loads hierarchy config from a JSON file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadFromEnv loads hierarchy config from environment or returns default
func LoadFromEnv() *Config {
	configPath := os.Getenv("HIERARCHY_CONFIG_PATH")
	if configPath != "" {
		if config, err := LoadFromFile(configPath); err == nil {
			return config
		}
	}

	// Check for preset
	preset := os.Getenv("HIERARCHY_PRESET")
	switch preset {
	case "ml-platform":
		return MLPlatformConfig()
	case "devops":
		return DevOpsConfig()
	default:
		return DefaultConfig()
	}
}

// GetLevel returns a level by name
func (c *Config) GetLevel(name string) *Level {
	for i := range c.Levels {
		if c.Levels[i].Name == name {
			return &c.Levels[i]
		}
	}
	return nil
}

// GetLevelByIndex returns a level by index
func (c *Config) GetLevelByIndex(index int) *Level {
	if index < 0 || index >= len(c.Levels) {
		return nil
	}
	return &c.Levels[index]
}

// GetParentLevel returns the parent level of a given level
func (c *Config) GetParentLevel(name string) *Level {
	for i, level := range c.Levels {
		if level.Name == name && i > 0 {
			return &c.Levels[i-1]
		}
	}
	return nil
}

// GetChildLevel returns the child level of a given level
func (c *Config) GetChildLevel(name string) *Level {
	for i, level := range c.Levels {
		if level.Name == name && i < len(c.Levels)-1 {
			return &c.Levels[i+1]
		}
	}
	return nil
}

// Depth returns the number of levels in the hierarchy
func (c *Config) Depth() int {
	return len(c.Levels)
}

// IsLeafLevel checks if a level is the leaf (resource-scoping) level
func (c *Config) IsLeafLevel(name string) bool {
	return name == c.LeafLevel
}

// NonRootLevels returns all levels except the root
func (c *Config) NonRootLevels() []Level {
	var levels []Level
	for _, level := range c.Levels {
		if !level.IsRoot {
			levels = append(levels, level)
		}
	}
	return levels
}
