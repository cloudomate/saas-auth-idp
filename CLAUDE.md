# CLAUDE.md - Project Context

This file provides context for Claude Code (or any AI assistant) working on this codebase.

## Project Overview

**SaaS Auth IDP** is a production-ready, multi-tenant authentication and authorization system designed as a foundation for building SaaS applications. It provides:

- A headless Go backend with REST APIs
- A React SDK (`@saas-starter/react`) for frontend integration
- Pluggable organizational hierarchies (tenant → workspace, or custom)
- Fine-grained authorization via OpenFGA (Zanzibar-style)
- OAuth 2.0 (Google, GitHub) and email/password authentication

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Frontend App (uses @saas-starter/react SDK)                │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│  Traefik Gateway (Port 4455) ─────► AuthZ Service (8002)    │
│  - Routes API requests              - JWT validation        │
│  - ForwardAuth middleware           - API key validation    │
│                                     - OpenFGA permission    │
└─────────────────────────┬───────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │ Backend  │    │ OpenFGA  │    │PostgreSQL│
    │ API:8000 │    │   :8081  │    │  :5432   │
    └──────────┘    └──────────┘    └──────────┘
```

## Directory Structure

```
saas-auth-idp/
├── backend/                    # Go REST API server
│   ├── cmd/api/               # Entry point (main.go)
│   └── internal/
│       ├── api/handlers/      # HTTP handlers (auth, tenant, workspace)
│       ├── api/middleware/    # Auth middleware
│       ├── config/            # Configuration loading
│       └── models/            # GORM database models
├── authz/                      # Go ForwardAuth service
│   ├── cmd/authz/             # Entry point
│   └── internal/
│       ├── auth/              # JWT & API key validation
│       ├── authz/             # OpenFGA client
│       └── handlers/          # Gate handler (ForwardAuth)
├── packages/react/             # React SDK (@saas-starter/react)
│   └── src/
│       ├── contexts/          # SaasAuthContext (main state)
│       ├── hooks/             # useSaasAuth, useTenant, useWorkspaces, useHierarchy
│       ├── components/        # ProtectedRoute, SocialLoginButtons
│       └── types.ts           # TypeScript interfaces
├── deploy/
│   ├── openfga/               # OpenFGA authorization model
│   ├── traefik/               # Traefik gateway config
│   ├── hierarchy.json         # Default hierarchy config
│   └── hierarchy-examples/    # Example hierarchies (ML platform, DevOps)
├── docs/                       # Documentation
├── docker-compose.yml          # Full stack orchestration
└── .env.example                # Environment variables template
```

## Key Technologies

| Component | Technology | Purpose |
|-----------|------------|---------|
| Backend API | Go + Gin | REST API, business logic |
| AuthZ Service | Go + Gin | ForwardAuth, permission checks |
| Database | PostgreSQL + GORM | User, tenant, workspace storage |
| Authorization | OpenFGA | Fine-grained permissions (Zanzibar) |
| API Gateway | Traefik | Routing, ForwardAuth integration |
| Frontend SDK | React + TypeScript | Hooks and components for auth |

## Development Commands

```bash
# Start all services
docker compose up -d

# Run backend locally (requires Postgres running)
cd backend && go run ./cmd/api

# Run authz service locally
cd authz && go run ./cmd/authz

# Build React SDK
cd packages/react && npm run build

# Run with dev mode (bypasses auth)
DEV_MODE=true docker compose up -d
```

## API Endpoints Summary

### Authentication (Public)
- `GET /api/v1/auth/social/:provider/login` - Get OAuth URL
- `POST /api/v1/auth/social/callback` - OAuth callback
- `POST /api/v1/auth/register` - Email registration
- `POST /api/v1/auth/verify-email` - Verify email
- `POST /api/v1/auth/login` - Email/password login
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password
- `GET /api/v1/auth/me` - Get current user (protected)

### Tenant (Protected)
- `GET /api/v1/tenant` - Get current tenant
- `GET /api/v1/tenant/plans` - List subscription plans
- `POST /api/v1/tenant/select-plan` - Select plan
- `POST /api/v1/tenant/setup` - Create organization
- `GET /api/v1/tenant/check-slug` - Check slug availability

### Workspaces (Protected)
- `GET /api/v1/workspaces` - List workspaces
- `POST /api/v1/workspaces` - Create workspace
- `GET /api/v1/workspaces/:id` - Get workspace
- `DELETE /api/v1/workspaces/:id` - Delete workspace
- `GET /api/v1/workspaces/:id/members` - List members
- `POST /api/v1/workspaces/:id/members` - Add member

## Key Concepts

### Multi-Tenancy
- **Tenant**: Root organization (company/team)
- **Workspace**: Project within a tenant
- **User**: Belongs to tenant, can access multiple workspaces

### Pluggable Hierarchies
The system supports custom organizational hierarchies via `deploy/hierarchy.json`:
- Default: `tenant → workspace`
- ML Platform: `tenant → team → project`
- DevOps: `tenant → environment → service`

### Authorization Model (OpenFGA)
Uses a generic "container" abstraction for any hierarchy level:
- Roles: `admin`, `member`, `viewer`
- Permissions inherit from parent containers
- Resources can be scoped to containers

## Environment Variables

Critical variables to configure:
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret for signing JWT tokens
- `API_KEY_SECRET` - Secret for API key validation
- `GOOGLE_CLIENT_ID/SECRET` - Google OAuth credentials
- `GITHUB_CLIENT_ID/SECRET` - GitHub OAuth credentials
- `OPENFGA_URL` - OpenFGA service URL
- `DEV_MODE` - Set to `true` to bypass auth in development

## Common Tasks

### Adding a new API endpoint
1. Add handler in `backend/internal/api/handlers/`
2. Register route in `backend/cmd/api/main.go`
3. Add middleware (RequireAuth, RequireTenant) as needed

### Adding a new React hook
1. Create hook in `packages/react/src/hooks/`
2. Export from `packages/react/src/index.ts`
3. Add TypeScript types to `packages/react/src/types.ts`

### Modifying the authorization model
1. Edit `deploy/openfga/model.fga`
2. Recreate OpenFGA store and upload new model
3. Update authz service if permission checks change

## Testing

```bash
# Backend tests
cd backend && go test ./...

# AuthZ tests
cd authz && go test ./...

# React SDK tests
cd packages/react && npm test
```

## Important Files

- `backend/internal/models/models.go` - Database schema definitions
- `backend/internal/api/handlers/auth.go` - Authentication logic
- `authz/internal/handlers/gate.go` - ForwardAuth handler
- `packages/react/src/contexts/SaasAuthContext.tsx` - Main auth context
- `deploy/openfga/model.fga` - Authorization model
- `deploy/hierarchy.json` - Hierarchy configuration
