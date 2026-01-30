# Sample API with ReBAC/ABAC Authorization

This sample API demonstrates how to implement fine-grained authorization using ReBAC (Relationship-Based Access Control) and ABAC (Attribute-Based Access Control) patterns.

## Overview

The API provides two resource types that showcase different authorization patterns:

### Documents (ReBAC)
Access is determined by **relationships** between users and documents.

```
Relationships:
- owner: Full control (read, write, delete, share)
- editor: Can read and write
- viewer: Can only read

Features:
- Document sharing with specific users
- Visibility levels (public, workspace, private)
- Permission inheritance from workspace
```

### Projects (ABAC)
Access is determined by **attributes** of users, resources, and context.

```
Attributes Evaluated:
- User: role (admin, member), platform admin status
- Resource: environment (production, staging, development)
- Resource: status (active, paused, archived)

Policies:
- Only admins can modify production projects
- Archived projects are read-only
- Paused projects cannot be deployed
- Owners can delete non-production projects
```

## Running the API

### Prerequisites

- Go 1.21+
- Running OpenFGA instance (optional - works with mock auth)

### Start the API

```bash
cd examples/sample-api

# Without OpenFGA (mock authorization)
go run main.go

# With OpenFGA
OPENFGA_URL=http://localhost:8081 OPENFGA_STORE_ID=your-store-id go run main.go
```

The API runs on port 8001 by default.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8001` | API port |
| `OPENFGA_URL` | `http://localhost:8081` | OpenFGA URL |
| `OPENFGA_STORE_ID` | - | OpenFGA store ID (required for real authz) |

## API Endpoints

### Documents (ReBAC Demo)

```bash
# List documents (filtered by visibility and shares)
GET /api/v1/documents

# Create document
POST /api/v1/documents
{
  "title": "My Document",
  "content": "Document content...",
  "visibility": "workspace"  // public, workspace, private
}

# Get document with permissions
GET /api/v1/documents/:id

# Update document (requires editor or owner)
PUT /api/v1/documents/:id
{
  "title": "Updated Title",
  "status": "published"
}

# Delete document (owner only)
DELETE /api/v1/documents/:id

# Share document (owner only)
POST /api/v1/documents/:id/share
{
  "user_id": "user-123",
  "role": "editor"  // editor, viewer
}

# Get user's permissions on document
GET /api/v1/documents/:id/permissions
```

### Projects (ABAC Demo)

```bash
# List projects with ABAC policies
GET /api/v1/projects
GET /api/v1/projects?environment=production

# Create project (production requires admin)
POST /api/v1/projects
{
  "name": "My Project",
  "description": "Project description",
  "environment": "development",  // production, staging, development
  "tags": ["backend", "api"]
}

# Get project with evaluated permissions
GET /api/v1/projects/:id

# Update project (ABAC policies apply)
PUT /api/v1/projects/:id
{
  "name": "Updated Name",
  "status": "active",
  "environment": "staging"
}

# Delete project (owner for non-prod, admin for prod)
DELETE /api/v1/projects/:id

# Deploy project (ABAC policies apply)
POST /api/v1/projects/:id/deploy
```

### Permission Check

```bash
# Direct permission check (for debugging)
POST /api/v1/check-permission
{
  "user": "user:user-123",
  "relation": "can_read",
  "object": "document:doc-456"
}
```

## ReBAC Pattern Explained

ReBAC determines access based on relationships:

```
┌─────────────────────────────────────────────────────────────┐
│                     Document: doc-1                          │
├─────────────────────────────────────────────────────────────┤
│ Relationships:                                               │
│   • owner: user-1                                           │
│   • editor: user-2                                          │
│   • viewer: user-3                                          │
│   • container: workspace-1                                  │
├─────────────────────────────────────────────────────────────┤
│ Derived Permissions:                                         │
│   • user-1: can_read ✓ can_write ✓ can_delete ✓ can_share ✓│
│   • user-2: can_read ✓ can_write ✓ can_delete ✗ can_share ✗│
│   • user-3: can_read ✓ can_write ✗ can_delete ✗ can_share ✗│
└─────────────────────────────────────────────────────────────┘
```

### OpenFGA Model for Documents

```fga
type document
  relations
    define container: [container]
    define owner: [user]
    define editor: [user]
    define viewer: [user]

    define can_read: viewer or editor or owner or can_read from container
    define can_write: editor or owner
    define can_delete: owner
    define can_share: owner
```

## ABAC Pattern Explained

ABAC evaluates policies based on attributes:

```
┌─────────────────────────────────────────────────────────────┐
│                      Policy Engine                           │
├─────────────────────────────────────────────────────────────┤
│ User Attributes:          Resource Attributes:              │
│   • user_id: user-1         • environment: production       │
│   • roles: [member]         • status: active                │
│   • is_admin: false         • owner_id: user-2              │
├─────────────────────────────────────────────────────────────┤
│ Policy Evaluation:                                           │
│                                                              │
│   Policy: "production_admin_only"                           │
│   Condition: resource.environment == "production"           │
│              AND user.is_admin == false                     │
│   Result: DENY write, deploy                                │
│                                                              │
│   Policy: "owner_can_delete"                                │
│   Condition: resource.owner_id != user.user_id              │
│              AND user.is_admin == false                     │
│   Result: DENY delete                                       │
├─────────────────────────────────────────────────────────────┤
│ Final Permissions:                                           │
│   can_read: ✓  can_write: ✗  can_delete: ✗  can_deploy: ✗  │
└─────────────────────────────────────────────────────────────┘
```

### ABAC Policies in Code

```go
func evaluateABACPolicies(user *UserContext, project *Project) map[string]bool {
    isAdmin := user.IsPlatformAdmin
    isOwner := project.OwnerID == user.UserID
    isProduction := project.Environment == "production"
    isArchived := project.Status == "archived"

    canWrite := true
    canDeploy := true

    // Policy: Archived projects are read-only
    if isArchived {
        canWrite = false
        canDeploy = false
    }

    // Policy: Production requires admin
    if isProduction && !isAdmin {
        canWrite = false
        canDeploy = false
    }

    return map[string]bool{
        "can_read":   true,
        "can_write":  canWrite,
        "can_delete": (isOwner || isAdmin) && !isArchived,
        "can_deploy": canDeploy,
    }
}
```

## Testing Authorization

### Test ReBAC (Documents)

```bash
# Create a private document as user-1
curl -X POST http://localhost:8001/api/v1/documents \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user-1" \
  -H "X-Workspace-ID: workspace-1" \
  -d '{"title": "Private Doc", "visibility": "private"}'

# Try to read as user-2 (should fail for private)
curl http://localhost:8001/api/v1/documents/DOC_ID \
  -H "X-User-ID: user-2" \
  -H "X-Workspace-ID: workspace-1"

# Share with user-2 as editor
curl -X POST http://localhost:8001/api/v1/documents/DOC_ID/share \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user-1" \
  -d '{"user_id": "user-2", "role": "editor"}'

# Now user-2 can read and write
curl http://localhost:8001/api/v1/documents/DOC_ID \
  -H "X-User-ID: user-2" \
  -H "X-Workspace-ID: workspace-1"
```

### Test ABAC (Projects)

```bash
# Create production project as admin
curl -X POST http://localhost:8001/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "X-User-ID: admin-1" \
  -H "X-Is-Platform-Admin: true" \
  -H "X-Workspace-ID: workspace-1" \
  -d '{"name": "Prod API", "environment": "production"}'

# Try to deploy as non-admin (should fail)
curl -X POST http://localhost:8001/api/v1/projects/PROJ_ID/deploy \
  -H "X-User-ID: user-1" \
  -H "X-Workspace-ID: workspace-1"
# Response: {"error": "policy_violation", "message": "Only administrators can deploy to production"}

# Deploy as admin (should succeed)
curl -X POST http://localhost:8001/api/v1/projects/PROJ_ID/deploy \
  -H "X-User-ID: admin-1" \
  -H "X-Is-Platform-Admin: true" \
  -H "X-Workspace-ID: workspace-1"
# Response: {"message": "Deployment initiated", ...}
```

## Integration with Main Auth System

When integrated with the main SaaS Auth IDP system:

1. **AuthZ Service** validates JWT and sets headers
2. **Sample API** reads headers via middleware
3. **OpenFGA** provides fine-grained permission checks

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│  Client  │───▶│ Traefik  │───▶│  AuthZ   │───▶│Sample API│
│          │    │ Gateway  │    │ Service  │    │          │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
                                     │               │
                                     ▼               ▼
                                ┌──────────┐   ┌──────────┐
                                │ OpenFGA  │   │  Store   │
                                │(ReBAC)   │   │(ABAC)    │
                                └──────────┘   └──────────┘
```

## Project Structure

```
sample-api/
├── main.go                     # Entry point, routing
├── go.mod
├── internal/
│   ├── authz/
│   │   └── openfga.go         # OpenFGA client wrapper
│   ├── handlers/
│   │   ├── documents.go       # ReBAC demo endpoints
│   │   └── projects.go        # ABAC demo endpoints
│   ├── middleware/
│   │   └── auth.go            # Header extraction
│   └── store/
│       ├── models.go          # Data models
│       └── memory.go          # In-memory store
└── README.md
```
