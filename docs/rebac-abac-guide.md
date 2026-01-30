# ReBAC and ABAC Authorization Guide

This guide explains the two main authorization patterns used in the system: **ReBAC** (Relationship-Based Access Control) and **ABAC** (Attribute-Based Access Control).

## Overview

| Pattern | Access Based On | Best For | Example |
|---------|----------------|----------|---------|
| **ReBAC** | Relationships between users and resources | Document sharing, team collaboration | "Alice is an editor of Document X" |
| **ABAC** | Attributes of users, resources, and context | Environment-based policies, compliance | "Only admins can deploy to production" |

## ReBAC (Relationship-Based Access Control)

### Concept

ReBAC determines access by examining **relationships** between subjects (users) and objects (resources).

```
User ──[relationship]──▶ Resource
```

### Key Components

1. **Subjects**: Users or groups
2. **Objects**: Resources (documents, projects, etc.)
3. **Relations**: Types of relationships (owner, editor, viewer)
4. **Permissions**: Derived from relations (can_read, can_write, can_delete)

### Example: Document Sharing

```
┌─────────────────────────────────────────────────────────────┐
│                    Document: "Q1 Roadmap"                   │
├─────────────────────────────────────────────────────────────┤
│ Relationships:                                               │
│                                                              │
│   Alice ──[owner]──▶ Document                               │
│   Bob ──[editor]──▶ Document                                │
│   Charlie ──[viewer]──▶ Document                            │
│   Workspace-1 ──[container]──▶ Document                     │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│ Derived Permissions:                                         │
│                                                              │
│   Alice:   can_read ✓  can_write ✓  can_delete ✓  can_share ✓│
│   Bob:     can_read ✓  can_write ✓  can_delete ✗  can_share ✗│
│   Charlie: can_read ✓  can_write ✗  can_delete ✗  can_share ✗│
└─────────────────────────────────────────────────────────────┘
```

### OpenFGA Model for ReBAC

```fga
type user

type document
  relations
    # Direct relationships
    define container: [container]
    define owner: [user]
    define editor: [user]
    define viewer: [user]

    # Derived permissions
    define can_read: viewer or editor or owner or can_read from container
    define can_write: editor or owner
    define can_delete: owner
    define can_share: owner
```

### Implementing ReBAC

#### 1. Creating Relationships

When a document is created:

```go
// User creates a document - they become the owner
fga.WriteTuple("user:alice", "owner", "document:doc-123")

// Document belongs to a workspace
fga.WriteTuple("container:workspace-1", "container", "document:doc-123")
```

When sharing a document:

```go
// Share with Bob as editor
fga.WriteTuple("user:bob", "editor", "document:doc-123")

// Share with Charlie as viewer
fga.WriteTuple("user:charlie", "viewer", "document:doc-123")
```

#### 2. Checking Permissions

```go
// Can Bob write to the document?
allowed, _ := fga.Check("user:bob", "can_write", "document:doc-123")
// Result: true (Bob is an editor)

// Can Charlie write to the document?
allowed, _ := fga.Check("user:charlie", "can_write", "document:doc-123")
// Result: false (Charlie is only a viewer)
```

#### 3. Permission Inheritance

Permissions can inherit from parent containers:

```
Workspace-1 (Alice: admin)
└── Document-1 (no direct owner)
    └── Alice can read via workspace inheritance
```

```fga
define can_read: viewer or editor or owner or can_read from container
```

### ReBAC Use Cases

- **Document Management**: Share documents with specific users
- **Project Collaboration**: Team members with different access levels
- **File Systems**: Folder/file permission inheritance
- **Social Networks**: Friends, followers, blocked users

---

## ABAC (Attribute-Based Access Control)

### Concept

ABAC determines access by evaluating **attributes** of the subject, resource, action, and context.

```
Policy(subject_attrs, resource_attrs, action, context) → ALLOW/DENY
```

### Key Components

1. **Subject Attributes**: User properties (role, department, clearance)
2. **Resource Attributes**: Resource properties (environment, status, classification)
3. **Action**: What the user wants to do (read, write, deploy)
4. **Context Attributes**: Environmental factors (time, location, IP)

### Example: Deployment Policies

```
┌─────────────────────────────────────────────────────────────┐
│                      Policy Engine                           │
├─────────────────────────────────────────────────────────────┤
│ Subject Attributes:        Resource Attributes:              │
│   • user_id: alice           • environment: production       │
│   • role: developer          • status: active                │
│   • department: engineering  • classification: internal      │
│                                                              │
│ Action: deploy                                               │
│                                                              │
│ Context:                                                     │
│   • time: 2:30 PM                                           │
│   • ip: 10.0.0.50 (internal)                                │
├─────────────────────────────────────────────────────────────┤
│ Policy Evaluation:                                           │
│                                                              │
│ Policy 1: "Production deployments require admin role"        │
│   Condition: resource.env == "production" AND                │
│              user.role != "admin"                            │
│   Result: DENY                                               │
│                                                              │
│ Policy 2: "No deployments after business hours"             │
│   Condition: time.hour > 18 OR time.hour < 9                │
│   Result: ALLOW (current time is within hours)              │
│                                                              │
│ Final Decision: DENY (Policy 1 denied)                      │
└─────────────────────────────────────────────────────────────┘
```

### ABAC Policy Examples

```go
// Policy 1: Production requires admin
func canDeployToProduction(user UserContext, project Project) bool {
    if project.Environment == "production" {
        return user.IsAdmin()
    }
    return true
}

// Policy 2: Archived resources are read-only
func canWrite(user UserContext, resource Resource) bool {
    if resource.Status == "archived" {
        return false
    }
    return true
}

// Policy 3: Only owner can delete non-production
func canDelete(user UserContext, project Project) bool {
    if project.Environment == "production" {
        return user.IsAdmin()
    }
    return project.OwnerID == user.ID || user.IsAdmin()
}

// Policy 4: Business hours only for production changes
func canModifyProduction(user UserContext, project Project, ctx Context) bool {
    if project.Environment != "production" {
        return true
    }
    hour := ctx.Time.Hour()
    return hour >= 9 && hour <= 18 && !ctx.IsWeekend()
}
```

### Implementing ABAC

#### 1. Define Policies

```go
type Policy struct {
    Name        string
    Description string
    Evaluate    func(subject, resource, context) bool
}

var policies = []Policy{
    {
        Name:        "production_admin_only",
        Description: "Only admins can modify production",
        Evaluate: func(user, project, ctx) bool {
            if project.Environment == "production" {
                return user.Role == "admin"
            }
            return true
        },
    },
    {
        Name:        "archived_read_only",
        Description: "Archived projects are read-only",
        Evaluate: func(user, project, ctx) bool {
            return project.Status != "archived"
        },
    },
}
```

#### 2. Evaluate Policies

```go
func EvaluatePermissions(user UserContext, project Project) map[string]bool {
    permissions := map[string]bool{
        "can_read":   true,
        "can_write":  true,
        "can_delete": true,
        "can_deploy": true,
    }

    // Apply each policy
    for _, policy := range policies {
        if !policy.Evaluate(user, project, context.Background()) {
            // Policy denied - restrict permissions
            switch policy.Name {
            case "production_admin_only":
                permissions["can_write"] = false
                permissions["can_deploy"] = false
            case "archived_read_only":
                permissions["can_write"] = false
                permissions["can_delete"] = false
                permissions["can_deploy"] = false
            }
        }
    }

    return permissions
}
```

#### 3. Enforce in Handlers

```go
func (h *Handler) Deploy(c *gin.Context) {
    user := GetUserContext(c)
    project := h.getProject(c.Param("id"))

    permissions := EvaluatePermissions(user, project)

    if !permissions["can_deploy"] {
        c.JSON(403, gin.H{
            "error":   "policy_violation",
            "message": "Deployment not allowed",
            "reason":  GetDenialReason(user, project, "deploy"),
        })
        return
    }

    // Proceed with deployment
    h.deployProject(project)
}
```

### ABAC Use Cases

- **Environment Controls**: Different rules for prod/staging/dev
- **Compliance**: Data classification, retention policies
- **Time-Based Access**: Business hours, maintenance windows
- **Location-Based**: IP restrictions, geo-fencing
- **Risk-Based**: Adaptive authentication based on risk score

---

## Combining ReBAC and ABAC

In practice, you often combine both patterns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Authorization Check                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. ReBAC Check: Does user have relationship?               │
│     └── user:alice ──[editor]──▶ document:doc-1             │
│     └── Result: has "editor" relationship                   │
│                                                              │
│  2. ABAC Check: Do attributes allow access?                 │
│     └── document.status != "archived" ✓                     │
│     └── document.classification <= user.clearance ✓         │
│     └── context.time within business_hours ✓                │
│     └── Result: attributes allow access                     │
│                                                              │
│  3. Final Decision: ALLOW (both checks passed)              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Example

```go
func CanWrite(user UserContext, document Document) bool {
    // ReBAC: Check relationship
    hasRelationship := fga.Check(
        fmt.Sprintf("user:%s", user.ID),
        "can_write",
        fmt.Sprintf("document:%s", document.ID),
    )

    if !hasRelationship {
        return false
    }

    // ABAC: Check attributes
    if document.Status == "archived" {
        return false
    }

    if document.Classification > user.ClearanceLevel {
        return false
    }

    return true
}
```

---

## Best Practices

### 1. Start Simple

Begin with ReBAC for basic sharing, add ABAC policies as needed:

```go
// Start with simple ownership
if resource.OwnerID == user.ID {
    return true
}

// Add sharing (ReBAC)
if hasRelationship(user, resource, "editor") {
    return true
}

// Add policy rules (ABAC)
if resource.Environment == "production" && !user.IsAdmin {
    return false
}
```

### 2. Explain Denials

Always provide clear reasons when access is denied:

```go
type AccessDenied struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Reason  string `json:"reason"`
    Policy  string `json:"policy,omitempty"`
}

// Response:
{
    "error": "access_denied",
    "message": "You cannot deploy this project",
    "reason": "Only administrators can deploy to production",
    "policy": "production_admin_only"
}
```

### 3. Audit Everything

Log all authorization decisions:

```go
logger.Info("authorization_decision",
    "user_id", user.ID,
    "action", "deploy",
    "resource", project.ID,
    "allowed", allowed,
    "policies_evaluated", []string{"production_admin_only", "archived_read_only"},
    "denial_reason", reason,
)
```

### 4. Cache Appropriately

- **ReBAC**: Cache relationship checks (with TTL)
- **ABAC**: Be careful caching attribute-based decisions (attributes change)

### 5. Test Authorization

```go
func TestProductionDeployPolicy(t *testing.T) {
    tests := []struct {
        name     string
        user     UserContext
        project  Project
        expected bool
    }{
        {
            name:     "admin can deploy to production",
            user:     UserContext{Role: "admin"},
            project:  Project{Environment: "production"},
            expected: true,
        },
        {
            name:     "member cannot deploy to production",
            user:     UserContext{Role: "member"},
            project:  Project{Environment: "production"},
            expected: false,
        },
        {
            name:     "member can deploy to staging",
            user:     UserContext{Role: "member"},
            project:  Project{Environment: "staging"},
            expected: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CanDeploy(tt.user, tt.project)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

---

## Example Applications

See the working examples in the `examples/` directory:

- **[sample-api/](../examples/sample-api/)**: Go backend demonstrating ReBAC and ABAC
- **[react-app/](../examples/react-app/)**: React frontend with permission-aware UI

### Running the Examples

```bash
# Terminal 1: Start the auth system
docker compose up -d

# Terminal 2: Start the sample API
cd examples/sample-api
go run main.go

# Terminal 3: Start the React app
cd examples/react-app
npm install
npm run dev
```

Then open http://localhost:3000 and navigate to:
- **Documents** page: ReBAC demo (sharing, visibility)
- **Projects** page: ABAC demo (environment policies, admin toggle)
