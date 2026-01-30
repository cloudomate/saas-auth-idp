# Authorization Guide

This guide explains the fine-grained authorization system built on OpenFGA.

## Overview

The system uses [OpenFGA](https://openfga.dev/), a Zanzibar-style authorization engine, for fine-grained access control. This provides:

- **Relationship-Based Access Control (ReBAC)**: Permissions based on relationships
- **Role-Based Access Control (RBAC)**: Roles with predefined permissions
- **Hierarchical Inheritance**: Permissions flow from parent to child containers

## Authorization Model

The authorization model is defined in `deploy/openfga/model.fga`:

```fga
model
  schema 1.1

type user

type platform
  relations
    define admin: [user]

type container
  relations
    # Parent relationship (for hierarchy traversal)
    define parent: [container]

    # Direct roles at this container level
    define admin: [user]
    define member: [user] or admin
    define viewer: [user] or member

    # Inherited permissions from parent
    define parent_admin: admin from parent
    define parent_member: member from parent

    # Effective permissions (direct + inherited)
    define can_manage: admin or parent_admin
    define can_write: member or can_manage or parent_member
    define can_read: viewer or can_write

type resource
  relations
    define container: [container]
    define owner: [user]

    define can_manage: owner or can_manage from container
    define can_write: can_write from container
    define can_read: can_read from container

type api_key
  relations
    define container: [container]
    define owner: [user]

    define can_use: owner
    define can_read: can_read from container
    define can_write: can_write from container
```

## Core Concepts

### 1. Containers

Containers are the generic abstraction for organizational hierarchy levels:

| Container Type | Example | Typical Roles |
|---------------|---------|---------------|
| Tenant | Organization | admin, member |
| Workspace | Project/Team | admin, member, viewer |
| Project | ML Project | admin, contributor, viewer |
| Environment | Production | admin, operator |

### 2. Roles

Each role has specific capabilities:

| Role | can_read | can_write | can_manage |
|------|----------|-----------|------------|
| viewer | ✓ | - | - |
| member | ✓ | ✓ | - |
| admin | ✓ | ✓ | ✓ |

**Permissions**:
- `can_read`: View containers and resources
- `can_write`: Create, update resources
- `can_manage`: Delete containers, manage members, admin operations

### 3. Permission Inheritance

Permissions flow down the hierarchy:

```
Tenant (Alice: admin)
└── Workspace (no direct roles)
    └── Alice has admin access via inheritance
```

```fga
# Relationships in OpenFGA:
user:alice#admin@container:tenant-1
container:workspace-1#parent@container:tenant-1

# Check: can Alice manage workspace-1?
check(user:alice, can_manage, container:workspace-1) → ALLOWED
  because: admin from parent (tenant-1)
```

## How Authorization Works

### Request Flow

```
1. Client Request
   GET /api/v1/workspaces/ws-123/resources
   Headers: Authorization: Bearer <jwt>

2. Traefik Gateway
   → Forward to AuthZ service

3. AuthZ Service (/gate)
   a. Validate JWT
   b. Extract: user_id, tenant_id
   c. Parse workspace from URL or header

4. OpenFGA Check
   check(user:user-123, can_read, container:ws-123)

5. Decision
   ALLOWED → Set headers, forward to backend
   DENIED → Return 403 Forbidden
```

### AuthZ Service Logic

```go
// gate.go (simplified)
func (h *GateHandler) Handle(c *gin.Context) {
    // 1. Validate token
    claims, err := h.jwtValidator.Validate(token)

    // 2. Skip auth for public routes
    if isPublicRoute(path) {
        return allow()
    }

    // 3. Platform admin bypasses checks
    if claims.IsPlatformAdmin {
        return allow()
    }

    // 4. Check OpenFGA for workspace access
    if workspaceID != "" {
        allowed := h.fgaClient.Check(
            fmt.Sprintf("user:%s", claims.UserID),
            "can_read",  // or can_write, can_manage based on method
            fmt.Sprintf("container:%s", workspaceID),
        )
        if !allowed {
            return deny(403, "forbidden")
        }
    }

    return allow()
}
```

## Setting Up Relationships

### When User Creates Tenant

```go
// Backend creates these relationships:
fga.Write([
    // User is admin of tenant
    {user: "user:123", relation: "admin", object: "container:tenant-456"},
])
```

### When Workspace is Created

```go
// Backend creates these relationships:
fga.Write([
    // Workspace parent is tenant
    {user: "container:tenant-456", relation: "parent", object: "container:workspace-789"},
    // Creator is admin of workspace
    {user: "user:123", relation: "admin", object: "container:workspace-789"},
])
```

### When Member is Added

```go
// Backend creates relationship:
fga.Write([
    {user: "user:new-member", relation: "member", object: "container:workspace-789"},
])
```

## API Key Authorization

API keys have limited permissions:

```fga
type api_key
  relations
    define container: [container]
    define owner: [user]

    define can_use: owner
    define can_read: can_read from container
    define can_write: can_write from container
    # Note: no can_manage - API keys can't perform admin operations
```

### API Key Flow

```
1. Request with API key
   Authorization: Bearer sk-key-123-secret...
   X-Workspace-ID: workspace-789

2. AuthZ validates API key
   - Verify signature
   - Check owner relationship
   - Check workspace relationship

3. Permission check
   check(api_key:key-123, can_write, container:workspace-789)
```

## Permission Checks by Endpoint

| Endpoint | Method | Required Permission |
|----------|--------|---------------------|
| `/workspaces` | GET | `can_read` on tenant |
| `/workspaces` | POST | `can_manage` on tenant |
| `/workspaces/:id` | GET | `can_read` on workspace |
| `/workspaces/:id` | DELETE | `can_manage` on workspace |
| `/workspaces/:id/members` | GET | `can_read` on workspace |
| `/workspaces/:id/members` | POST | `can_manage` on workspace |
| `/resources` | GET | `can_read` on workspace |
| `/resources` | POST | `can_write` on workspace |
| `/resources/:id` | DELETE | `can_manage` on workspace |

## Checking Permissions in Code

### Backend (Go)

```go
// Using OpenFGA client
func (h *Handler) DeleteWorkspace(c *gin.Context) {
    userID := c.GetString("user_id")
    workspaceID := c.Param("id")

    // Check permission
    allowed, err := h.fga.Check(
        fmt.Sprintf("user:%s", userID),
        "can_manage",
        fmt.Sprintf("container:%s", workspaceID),
    )

    if !allowed {
        c.JSON(403, gin.H{"error": "forbidden"})
        return
    }

    // Proceed with deletion
}
```

### Frontend (React)

The SDK doesn't expose permission checks directly. Instead, handle 403 responses:

```tsx
const { createWorkspace } = useWorkspaces()

const handleCreate = async () => {
    try {
        await createWorkspace(name, slug)
    } catch (err) {
        if (err.status === 403) {
            setError("You don't have permission to create workspaces")
        }
    }
}
```

## Custom Authorization Logic

### Adding Custom Roles

1. Update the hierarchy configuration:

```json
{
  "levels": [
    {
      "name": "workspace",
      "roles": ["admin", "editor", "viewer"]
    }
  ]
}
```

2. The OpenFGA model already supports custom roles through the generic `member` relation. Map your roles:
   - `editor` → `member` (can_write)
   - `viewer` → `viewer` (can_read)

### Resource-Level Permissions

For fine-grained resource permissions:

```fga
type resource
  relations
    define container: [container]
    define owner: [user]
    define editor: [user]
    define viewer: [user]

    define can_manage: owner or can_manage from container
    define can_write: editor or can_manage
    define can_read: viewer or can_write
```

Create relationships when resources are created:

```go
fga.Write([
    {user: "user:123", relation: "owner", object: "resource:doc-456"},
    {user: "container:workspace-789", relation: "container", object: "resource:doc-456"},
])
```

## OpenFGA Management

### Creating a Store

```bash
# Create store
curl -X POST http://localhost:8081/stores \
  -H "Content-Type: application/json" \
  -d '{"name": "saas-starter"}'

# Response contains store_id
```

### Uploading Authorization Model

```bash
# Upload model
curl -X POST http://localhost:8081/stores/{store_id}/authorization-models \
  -H "Content-Type: application/json" \
  -d @deploy/openfga/model.json

# Set OPENFGA_STORE_ID in .env
```

### Viewing Relationships

```bash
# List tuples
curl "http://localhost:8081/stores/{store_id}/read" \
  -H "Content-Type: application/json" \
  -d '{}'
```

### OpenFGA Playground

Access the OpenFGA playground at `http://localhost:3000` (when running via Docker Compose) to:
- Visualize the authorization model
- Test permission checks
- Debug relationship queries

## Best Practices

### 1. Principle of Least Privilege

Assign the minimum role needed:
- Use `viewer` for read-only access
- Use `member` for contributors
- Reserve `admin` for managers

### 2. Use Inheritance

Don't duplicate permissions. Let inheritance handle it:

```
# Good: Admin at tenant level
user:alice#admin@container:tenant-1

# Bad: Duplicating at every level
user:alice#admin@container:tenant-1
user:alice#admin@container:workspace-1
user:alice#admin@container:workspace-2
```

### 3. Audit Permission Changes

Log all relationship changes:

```go
func (h *Handler) AddMember(c *gin.Context) {
    // ... add member logic ...

    // Log the change
    h.logger.Info("member_added",
        "user_id", newMemberID,
        "workspace_id", workspaceID,
        "role", role,
        "added_by", currentUserID,
    )
}
```

### 4. Handle Permission Denials Gracefully

```tsx
function WorkspaceSettings() {
    const [canManage, setCanManage] = useState(false)

    useEffect(() => {
        // Try to fetch settings
        api.get('/workspace/settings')
            .then(() => setCanManage(true))
            .catch(err => {
                if (err.status === 403) {
                    setCanManage(false)
                }
            })
    }, [])

    if (!canManage) {
        return <div>You don't have permission to manage this workspace.</div>
    }

    return <SettingsForm />
}
```

## Troubleshooting

### Permission Denied When Expected

1. Check relationship exists:
```bash
curl "http://localhost:8081/stores/{store_id}/read" \
  -d '{"tuple_key": {"user": "user:123", "object": "container:workspace-456"}}'
```

2. Verify parent relationship for inheritance:
```bash
curl "http://localhost:8081/stores/{store_id}/read" \
  -d '{"tuple_key": {"relation": "parent", "object": "container:workspace-456"}}'
```

3. Test the check directly:
```bash
curl "http://localhost:8081/stores/{store_id}/check" \
  -d '{"tuple_key": {"user": "user:123", "relation": "can_read", "object": "container:workspace-456"}}'
```

### Stale Permissions

If permissions seem stale:
1. OpenFGA caches results - wait a few seconds
2. Verify the write operation succeeded
3. Check for conflicting relationships

### Model Updates

When updating the authorization model:
1. Create a new model version
2. Migrate existing relationships if needed
3. Update the store's active model
