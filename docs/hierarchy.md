# Hierarchy Configuration Guide

This guide explains the pluggable hierarchy system that allows you to customize your organizational structure.

## Overview

The system supports flexible organizational hierarchies through a configuration-driven approach. Instead of hardcoding "Tenant → Workspace", you can define any hierarchy:

- **Default**: Tenant → Workspace
- **ML Platform**: Organization → Team → Project
- **DevOps**: Organization → Environment → Service
- **Enterprise**: Company → Division → Department → Team

## Configuration File

The hierarchy is defined in `deploy/hierarchy.json`:

```json
{
  "root_level": "tenant",
  "leaf_level": "workspace",
  "levels": [
    {
      "name": "tenant",
      "display_name": "Organization",
      "plural": "organizations",
      "url_path": "tenant",
      "roles": ["admin", "member"],
      "is_root": true
    },
    {
      "name": "workspace",
      "display_name": "Workspace",
      "plural": "workspaces",
      "url_path": "workspaces",
      "roles": ["admin", "member", "viewer"],
      "is_root": false
    }
  ]
}
```

## Configuration Schema

### Top-Level Properties

| Property | Type | Description |
|----------|------|-------------|
| `root_level` | string | Name of the top-level container (e.g., "tenant") |
| `leaf_level` | string | Name of the bottom-level container (e.g., "workspace") |
| `levels` | array | Array of hierarchy level definitions |

### Level Properties

| Property | Type | Description |
|----------|------|-------------|
| `name` | string | Internal identifier (used in code/DB) |
| `display_name` | string | User-friendly name |
| `plural` | string | Plural form for display |
| `url_path` | string | API URL path segment |
| `roles` | string[] | Available roles at this level |
| `is_root` | boolean | Whether this is the root level |

## Example Hierarchies

### Default (2-Level)

```
Organization (tenant)
└── Workspace (workspace)
```

```json
{
  "root_level": "tenant",
  "leaf_level": "workspace",
  "levels": [
    {
      "name": "tenant",
      "display_name": "Organization",
      "plural": "organizations",
      "url_path": "tenant",
      "roles": ["admin", "member"],
      "is_root": true
    },
    {
      "name": "workspace",
      "display_name": "Workspace",
      "plural": "workspaces",
      "url_path": "workspaces",
      "roles": ["admin", "member", "viewer"],
      "is_root": false
    }
  ]
}
```

### ML Platform (3-Level)

```
Organization (tenant)
├── ML Team (team)
│   ├── Model Training Project (project)
│   └── Data Pipeline Project (project)
└── Data Science Team (team)
    └── Analytics Project (project)
```

```json
{
  "root_level": "tenant",
  "leaf_level": "project",
  "levels": [
    {
      "name": "tenant",
      "display_name": "Organization",
      "plural": "organizations",
      "url_path": "tenant",
      "roles": ["admin", "member"],
      "is_root": true
    },
    {
      "name": "team",
      "display_name": "Team",
      "plural": "teams",
      "url_path": "teams",
      "roles": ["lead", "member"],
      "is_root": false
    },
    {
      "name": "project",
      "display_name": "Project",
      "plural": "projects",
      "url_path": "projects",
      "roles": ["admin", "contributor", "viewer"],
      "is_root": false
    }
  ]
}
```

### DevOps Platform (3-Level)

```
Organization (tenant)
├── Production (environment)
│   ├── API Service (service)
│   └── Worker Service (service)
├── Staging (environment)
│   └── API Service (service)
└── Development (environment)
    └── API Service (service)
```

```json
{
  "root_level": "tenant",
  "leaf_level": "service",
  "levels": [
    {
      "name": "tenant",
      "display_name": "Organization",
      "plural": "organizations",
      "url_path": "tenant",
      "roles": ["admin", "member"],
      "is_root": true
    },
    {
      "name": "environment",
      "display_name": "Environment",
      "plural": "environments",
      "url_path": "environments",
      "roles": ["admin", "operator"],
      "is_root": false
    },
    {
      "name": "service",
      "display_name": "Service",
      "plural": "services",
      "url_path": "services",
      "roles": ["admin", "developer", "viewer"],
      "is_root": false
    }
  ]
}
```

## Using Hierarchies in the SDK

### useHierarchy Hook

The `useHierarchy` hook provides full access to hierarchy operations:

```tsx
import { useHierarchy } from '@saas-starter/react'

function HierarchyNavigator() {
  const {
    config,
    rootLevel,
    leafLevel,
    currentRoot,
    currentLeaf,
    listContainers,
    createContainer,
    setCurrentContainer,
  } = useHierarchy()

  // Display hierarchy levels
  return (
    <div>
      <h2>Hierarchy: {config?.levels.map(l => l.display_name).join(' → ')}</h2>

      {config?.levels.map(level => (
        <LevelSection
          key={level.name}
          level={level}
          onSelect={(container) => setCurrentContainer(level.name, container)}
        />
      ))}
    </div>
  )
}
```

### Dynamic Level Navigation

```tsx
function LevelBrowser({ levelName, parentId }) {
  const { getLevel, listContainers, createContainer } = useHierarchy()
  const [containers, setContainers] = useState([])

  const level = getLevel(levelName)

  useEffect(() => {
    listContainers(levelName, parentId).then(setContainers)
  }, [levelName, parentId])

  const handleCreate = async (name, slug) => {
    const container = await createContainer(levelName, name, slug, parentId)
    setContainers([...containers, container])
  }

  return (
    <div>
      <h3>{level.display_name}s</h3>
      <ul>
        {containers.map(c => (
          <li key={c.id}>
            {c.display_name}
            {/* Navigate to children */}
            {level.name !== config.leaf_level && (
              <Link to={`/${level.url_path}/${c.id}`}>
                View {getNextLevel(level).plural}
              </Link>
            )}
          </li>
        ))}
      </ul>
      <CreateContainerForm onSubmit={handleCreate} />
    </div>
  )
}
```

### useContainers Hook

For simpler use cases, `useContainers` works with a specific level:

```tsx
import { useContainers } from '@saas-starter/react'

function ProjectList() {
  const {
    containers: projects,
    currentContainer,
    setCurrentContainer,
    create,
    delete: deleteProject,
    level,
  } = useContainers('project')

  return (
    <div>
      <h2>{level.plural}</h2>
      {projects.map(project => (
        <ProjectCard
          key={project.id}
          project={project}
          isSelected={currentContainer?.id === project.id}
          onSelect={() => setCurrentContainer(project)}
          onDelete={() => deleteProject(project.id)}
        />
      ))}
    </div>
  )
}
```

## API Endpoints

The hierarchy generates dynamic API endpoints based on configuration:

### List Containers

```
GET /api/v1/{url_path}?parent_id={parent_id}
```

Examples:
- `GET /api/v1/workspaces` - List workspaces
- `GET /api/v1/teams?parent_id=tenant-123` - List teams in tenant
- `GET /api/v1/projects?parent_id=team-456` - List projects in team

### Create Container

```
POST /api/v1/{url_path}
{
  "name": "Container Name",
  "slug": "container-slug",
  "parent_id": "parent-container-id"
}
```

### Get Container

```
GET /api/v1/{url_path}/{id_or_slug}
```

### Delete Container

```
DELETE /api/v1/{url_path}/{id}
```

### List Members

```
GET /api/v1/{url_path}/{id}/members
```

### Add Member

```
POST /api/v1/{url_path}/{id}/members
{
  "email": "user@example.com",
  "role": "member"
}
```

## Database Schema

Containers are stored in a generic `resource_containers` table:

```sql
CREATE TABLE resource_containers (
    id UUID PRIMARY KEY,
    level VARCHAR(50) NOT NULL,      -- "tenant", "workspace", "project", etc.
    slug VARCHAR(100) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES resource_containers(id),
    root_id UUID NOT NULL,           -- Always points to tenant
    path VARCHAR(1000),              -- Materialized path: "tenant-1/team-2/project-3"
    depth INTEGER NOT NULL,          -- 0 for root, 1 for first child, etc.
    is_active BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,

    UNIQUE(level, slug, parent_id)
);
```

## Authorization Integration

The hierarchy integrates with OpenFGA through parent relationships:

```fga
type container
  relations
    define parent: [container]
    define admin: [user]
    define member: [user] or admin
    define viewer: [user] or member

    define parent_admin: admin from parent
    define parent_member: member from parent

    define can_manage: admin or parent_admin
    define can_write: member or can_manage or parent_member
    define can_read: viewer or can_write
```

When containers are created, relationships are established:

```go
// Creating a project under a team
fga.Write([
    // Project's parent is the team
    {user: "container:team-456", relation: "parent", object: "container:project-789"},
    // Creator is admin
    {user: "user:123", relation: "admin", object: "container:project-789"},
])
```

This enables permission inheritance:
- Team admin → Project admin (via parent_admin)
- Team member → Project member (via parent_member)

## Customization Examples

### Adding Metadata Fields

Use the `metadata` JSONB field for level-specific data:

```tsx
// Create workspace with metadata
await createContainer('workspace', 'Production', 'production', tenantId, {
  environment: 'production',
  region: 'us-east-1',
})

// Access metadata
const workspace = await getContainer('workspace', 'production')
console.log(workspace.metadata.environment) // 'production'
```

### Custom Roles per Level

Define different roles for each level:

```json
{
  "levels": [
    {
      "name": "environment",
      "roles": ["admin", "operator", "viewer"]
    },
    {
      "name": "service",
      "roles": ["admin", "developer", "tester", "viewer"]
    }
  ]
}
```

### Restricting Hierarchy Depth

The backend can enforce depth limits:

```go
func (h *Handler) CreateContainer(c *gin.Context) {
    parentDepth := parent.Depth
    maxDepth := len(h.hierarchyConfig.Levels) - 1

    if parentDepth >= maxDepth {
        c.JSON(400, gin.H{"error": "max_depth_exceeded"})
        return
    }
}
```

## Migrating Hierarchies

### Adding a Level

1. Update `hierarchy.json` with new level
2. Run migration to add existing containers to new level
3. Update UI to show new level

### Removing a Level

1. Migrate children to parent level
2. Delete containers at removed level
3. Update `hierarchy.json`
4. Update UI

### Example Migration Script

```sql
-- Adding "team" level between tenant and workspace

-- 1. Create default teams for each tenant
INSERT INTO resource_containers (id, level, slug, display_name, parent_id, root_id, depth)
SELECT
    gen_random_uuid(),
    'team',
    'default-team',
    'Default Team',
    id,
    id,
    1
FROM resource_containers
WHERE level = 'tenant';

-- 2. Update workspaces to point to default teams
UPDATE resource_containers ws
SET
    parent_id = t.id,
    depth = 2,
    path = CONCAT(t.path, '/', ws.slug)
FROM resource_containers t
WHERE ws.level = 'workspace'
  AND t.level = 'team'
  AND t.parent_id = ws.parent_id;
```

## Best Practices

### 1. Keep Hierarchies Shallow

Limit to 3-4 levels maximum. Deep hierarchies are harder to navigate and manage.

### 2. Use Meaningful Names

Choose `display_name` and `plural` that make sense to users:
- Good: "Project", "Projects"
- Bad: "Container Level 3", "Container Level 3s"

### 3. Consider URL Structure

The `url_path` becomes part of API URLs:
- `/api/v1/projects` (good)
- `/api/v1/project-containers` (verbose)

### 4. Plan Roles Carefully

Define roles that map to real responsibilities:
- `admin` - Full control
- `member/contributor/developer` - Create and modify
- `viewer/reader` - Read-only

### 5. Document Your Hierarchy

Add a comment or documentation explaining your hierarchy choice:

```json
{
  "_comment": "ML Platform hierarchy: Organization → Team → Project",
  "root_level": "tenant",
  "leaf_level": "project",
  "levels": [...]
}
```
