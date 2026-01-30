# Architecture Overview

This document explains the system architecture, components, and how they interact.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           YOUR APPLICATION                                   │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  import { SaasAuthProvider, useSaasAuth } from '@saas-starter/react'   │ │
│  │                                                                         │ │
│  │  <SaasAuthProvider apiUrl="https://api.yourapp.com">                   │ │
│  │    <YourApp />                                                         │ │
│  │  </SaasAuthProvider>                                                   │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         TRAEFIK GATEWAY (4455)                               │
│                    ┌──────────────────────────────┐                         │
│                    │      ForwardAuth             │                         │
│                    │   ┌──────────────────────┐   │                         │
│                    │   │  AuthZ Service       │   │                         │
│                    │   │  - JWT Validation    │   │                         │
│                    │   │  - API Key Validation│   │                         │
│                    │   │  - OpenFGA Check     │   │                         │
│                    │   └──────────────────────┘   │                         │
│                    └──────────────────────────────┘                         │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
           ┌─────────────────────────┼─────────────────────────┐
           ▼                         ▼                         ▼
┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│   Backend API    │    │   OpenFGA        │    │   PostgreSQL     │
│   (Port 8000)    │    │   (Port 8081)    │    │   (Port 5432)    │
│                  │    │                  │    │                  │
│ - Auth handlers  │    │ - Permissions    │    │ - Users          │
│ - Tenant mgmt    │    │ - Relationships  │    │ - Tenants        │
│ - Workspace mgmt │    │ - RBAC/ReBAC     │    │ - Workspaces     │
│ - Business logic │    │                  │    │ - Subscriptions  │
└──────────────────┘    └──────────────────┘    └──────────────────┘
```

## Components

### 1. API Gateway (Traefik)

**Port**: 4455 (external), 8080 (dashboard)

Traefik serves as the API gateway and handles:
- **Routing**: Directs requests to appropriate backend services
- **ForwardAuth**: Validates every request through the AuthZ service
- **SSL/TLS**: Handles HTTPS termination in production

Configuration: `deploy/traefik/`

### 2. AuthZ Service (ForwardAuth)

**Port**: 8002 (internal only)

A lightweight Go service that validates every incoming request:

```
Request Flow:
1. Client sends request with Authorization header
2. Traefik intercepts and calls AuthZ service (/gate)
3. AuthZ validates JWT or API key
4. AuthZ checks OpenFGA for permissions (if workspace-scoped)
5. AuthZ sets response headers (X-User-ID, X-Tenant-ID, etc.)
6. Traefik forwards request to backend with enriched headers
```

**Key Features**:
- JWT token validation
- API key validation (sk-* prefix)
- OpenFGA permission checks
- Dev mode bypass for local development

Source: `authz/internal/handlers/gate.go`

### 3. Backend API

**Port**: 8000 (internal only)

The main business logic server built with Go and Gin:

| Module | Path | Purpose |
|--------|------|---------|
| Auth | `/api/v1/auth/*` | OAuth, email/password, JWT |
| Tenant | `/api/v1/tenant/*` | Organization management |
| Workspaces | `/api/v1/workspaces/*` | Workspace CRUD, members |
| Hierarchy | `/api/v1/hierarchy` | Hierarchy configuration |

**Middleware Stack**:
1. CORS - Cross-origin request handling
2. RequireAuth - JWT validation (protected routes)
3. RequireTenant - Tenant existence check
4. RequireTenantAdmin - Admin-only operations

Source: `backend/internal/api/`

### 4. PostgreSQL Database

**Port**: 5432

Stores all persistent data:

```
┌─────────────────┐     ┌─────────────────┐
│     users       │     │    tenants      │
├─────────────────┤     ├─────────────────┤
│ id              │     │ id              │
│ email           │────▶│ admin_user_id   │
│ name            │     │ slug            │
│ auth_provider   │     │ display_name    │
│ is_tenant_admin │     │ subscription_id │
│ admin_of_tenant │◀────│                 │
└─────────────────┘     └─────────────────┘
        │                       │
        │                       │
        ▼                       ▼
┌─────────────────┐     ┌─────────────────┐
│  memberships    │     │   workspaces    │
├─────────────────┤     ├─────────────────┤
│ user_id         │     │ id              │
│ workspace_id    │────▶│ tenant_id       │
│ role            │     │ slug            │
└─────────────────┘     │ display_name    │
                        │ is_default      │
                        └─────────────────┘
```

Source: `backend/internal/models/models.go`

### 5. OpenFGA (Authorization)

**Port**: 8081

Zanzibar-style authorization engine for fine-grained permissions:

```
Authorization Model:

type user

type container  # tenant, workspace, project, etc.
  relations
    define parent: [container]
    define admin: [user]
    define member: [user] or admin
    define viewer: [user] or member
    define can_manage: admin or (admin from parent)
    define can_write: member or can_manage or (member from parent)
    define can_read: viewer or can_write
```

**Key Concepts**:
- **Containers**: Generic abstraction for any hierarchy level
- **Inheritance**: Permissions flow from parent to child containers
- **Roles**: admin > member > viewer

Source: `deploy/openfga/model.fga`

### 6. React SDK

**Package**: `@saas-starter/react`

Client-side library for frontend integration:

```
┌─────────────────────────────────────────┐
│         SaasAuthProvider                │
│  ┌───────────────────────────────────┐  │
│  │        SaasAuthContext            │  │
│  │  - User state                     │  │
│  │  - Tenant state                   │  │
│  │  - Workspace state                │  │
│  │  - API client                     │  │
│  └───────────────────────────────────┘  │
│              │                          │
│   ┌──────────┼──────────┬─────────┐    │
│   ▼          ▼          ▼         ▼    │
│ useSaasAuth useTenant useWorkspaces    │
│                        useHierarchy    │
└─────────────────────────────────────────┘
```

Source: `packages/react/src/`

## Request Flow

### Authenticated Request

```
1. Client: GET /api/v1/workspaces
   Headers: Authorization: Bearer <jwt>

2. Traefik: Forward to AuthZ service
   POST http://authz:8002/gate

3. AuthZ: Validate JWT
   - Decode token
   - Verify signature with JWT_SECRET
   - Check expiration
   - Extract claims (user_id, tenant_id)

4. AuthZ: Check OpenFGA (if workspace-scoped)
   - Query: check(user:123, viewer, container:workspace-456)
   - Returns: allowed/denied

5. AuthZ: Return to Traefik
   Headers:
     X-User-ID: 123
     X-Tenant-ID: abc
     X-Workspace-ID: 456
     X-Is-Platform-Admin: false

6. Traefik: Forward to Backend API
   GET http://api:8000/api/v1/workspaces
   + AuthZ headers

7. Backend: Process request
   - Read X-User-ID, X-Tenant-ID from headers
   - Execute business logic
   - Return response

8. Client: Receive response
```

### OAuth Login Flow

```
1. Client: loginWithGoogle()
   GET /api/v1/auth/social/google/login

2. Backend: Generate OAuth URL
   - Create state token (CSRF protection)
   - Store state in database
   - Return auth_url

3. Client: Redirect to Google
   User authorizes application

4. Google: Redirect to callback
   /api/v1/auth/social/callback?code=xxx&state=yyy

5. Client: POST /api/v1/auth/social/callback
   { code: "xxx", state: "yyy" }

6. Backend: Exchange code
   - Verify state token
   - Exchange code for access token
   - Fetch user info from Google
   - Create/update user in database
   - Generate JWT

7. Client: Store JWT
   - Save to localStorage
   - Update auth context
```

## Data Flow

### Multi-Tenant Isolation

```
Tenant A (tenant-a)              Tenant B (tenant-b)
├── Workspace 1                  ├── Workspace 3
│   ├── User Alice (admin)       │   ├── User Charlie (admin)
│   └── User Bob (member)        │   └── User Diana (viewer)
└── Workspace 2                  └── Workspace 4
    └── User Alice (admin)           └── User Charlie (admin)

- Users can only see workspaces in their tenant
- Permissions checked via OpenFGA
- API keys scoped to workspace level
```

### Permission Inheritance

```
Tenant (admin: Alice)
└── Workspace (no direct admin)
    └── Alice has admin via inheritance

Container hierarchy in OpenFGA:
- container:workspace#parent@container:tenant
- user:alice#admin@container:tenant
- Check: user:alice#can_manage@container:workspace → ALLOWED
```

## Scalability Considerations

### Horizontal Scaling

```
                    Load Balancer
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
    ┌─────────┐     ┌─────────┐     ┌─────────┐
    │ API Pod │     │ API Pod │     │ API Pod │
    └─────────┘     └─────────┘     └─────────┘
         │               │               │
         └───────────────┼───────────────┘
                         ▼
                   ┌──────────┐
                   │ Postgres │ (Primary)
                   └──────────┘
```

- Backend API is stateless - scale horizontally
- AuthZ service is stateless - scale horizontally
- PostgreSQL requires read replicas for scale
- OpenFGA can be clustered with shared storage

### Caching Strategy

- JWT validation: Token claims cached in memory
- User/tenant data: Consider Redis for session data
- OpenFGA: Built-in caching for permission checks

## Security Model

### Authentication Layers

1. **API Gateway**: TLS termination, rate limiting
2. **ForwardAuth**: Token validation, permission checks
3. **Backend**: Business logic authorization

### Token Security

- JWT signed with HS256 algorithm
- 24-hour expiration
- Contains: user_id, email, tenant_id, is_tenant_admin
- API keys use HMAC-SHA256 with separate secret

### Data Protection

- Passwords hashed with bcrypt
- OAuth state tokens for CSRF protection
- Email verification tokens expire in 24 hours
- Password reset tokens expire in 1 hour
