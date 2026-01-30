# SaaS Starter Kit

A production-ready, multi-tenant SaaS authentication and authorization system with a headless Go backend and React SDK.

## Architecture

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
│   Casdoor (IdP)  │    │   API Server     │    │   OpenFGA        │
│   - OIDC/OAuth   │    │   - REST API     │    │   - Permissions  │
│   - User Mgmt    │    │   - Business     │    │   - Relationships│
│   - SSO Config   │    │     Logic        │    │   - RBAC/ReBAC   │
└──────────────────┘    └──────────────────┘    └──────────────────┘
           │                         │                         │
           └─────────────────────────┼─────────────────────────┘
                                     ▼
                         ┌──────────────────┐
                         │   PostgreSQL     │
                         │   - Users        │
                         │   - Tenants      │
                         │   - Workspaces   │
                         └──────────────────┘
```

## Features

- **Multi-Tenant Architecture**: Tenant → Workspace → User hierarchy
- **Identity Provider (Casdoor)**: OIDC, OAuth 2.0, SSO configuration per tenant
- **Fine-Grained Authorization (OpenFGA)**: Zanzibar-style permissions (RBAC + ReBAC)
- **API Gateway (Traefik)**: ForwardAuth middleware for authentication
- **API Keys**: Workspace-scoped keys with hash-based validation
- **React SDK**: Hooks and components for easy integration
- **Subscription Plans**: Basic, Advanced, Enterprise tiers

## Quick Start

### 1. Start the Backend

```bash
# Clone the repository
git clone https://github.com/yourusername/saas-starter-kit.git
cd saas-starter-kit

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Start all services
docker compose up -d
```

### Services & Ports

| Service | Port | Description |
|---------|------|-------------|
| Traefik | 4455 | API Gateway (main entry point) |
| API | 8000 | Go backend (internal) |
| AuthZ | 8002 | ForwardAuth service (internal) |
| Casdoor | 8085 | Identity Provider |
| OpenFGA | 8081 | Authorization service |
| PostgreSQL | 5432 | Database |

### 2. Install the React SDK

```bash
npm install @saas-starter/react
# or
yarn add @saas-starter/react
```

### 3. Integrate in Your App

```tsx
import { SaasAuthProvider, useSaasAuth } from '@saas-starter/react'

function App() {
  return (
    <SaasAuthProvider
      apiUrl="http://localhost:4455"
      storagePrefix="myapp"
    >
      <YourApp />
    </SaasAuthProvider>
  )
}

function YourApp() {
  const {
    user,
    isAuthenticated,
    isLoading,
    loginWithGoogle,
    loginWithGithub,
    logout
  } = useSaasAuth()

  if (isLoading) return <div>Loading...</div>

  if (!isAuthenticated) {
    return (
      <div>
        <button onClick={() => loginWithGoogle()}>
          Sign in with Google
        </button>
      </div>
    )
  }

  return (
    <div>
      <p>Welcome, {user.name}!</p>
      <button onClick={logout}>Logout</button>
    </div>
  )
}
```

## Project Structure

```
saas-starter-kit/
├── backend/                    # Go API server
│   ├── cmd/api/               # Entry point
│   └── internal/
│       ├── api/handlers/      # HTTP handlers
│       ├── api/middleware/    # Auth middleware
│       ├── config/            # Configuration
│       └── models/            # Database models
├── authz/                      # AuthZ service (ForwardAuth)
│   ├── cmd/authz/             # Entry point
│   └── internal/
│       ├── auth/              # JWT & API key validation
│       ├── authz/             # OpenFGA client
│       └── handlers/          # Gate handler
├── packages/react/             # React SDK
│   └── src/
│       ├── contexts/          # SaasAuthContext
│       ├── hooks/             # useSaasAuth, useTenant, useWorkspaces
│       └── components/        # ProtectedRoute, SocialLoginButtons
├── deploy/
│   ├── casdoor/               # IdP configuration
│   ├── openfga/               # Authorization model
│   ├── traefik/               # Gateway configuration
│   └── postgres/              # Database init
├── docker-compose.yml          # Full stack deployment
└── .env.example                # Environment template
```

## API Reference

### REST API (via Gateway)

All API requests go through `http://localhost:4455` with authentication:

```bash
# With JWT token
curl -H "Authorization: Bearer <jwt_token>" \
     http://localhost:4455/api/v1/workspaces

# With API key
curl -H "Authorization: Bearer sk-<key_id>-<secret>" \
     -H "X-Workspace-ID: <workspace_uuid>" \
     http://localhost:4455/api/v1/workspaces
```

### Authentication Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/auth/social/:provider/login` | Get OAuth URL |
| POST | `/api/v1/auth/social/callback` | OAuth callback |
| POST | `/api/v1/auth/register` | Email registration |
| POST | `/api/v1/auth/verify-email` | Verify email |
| POST | `/api/v1/auth/login` | Email login |
| GET | `/api/v1/auth/me` | Get current user |

### Tenant Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/tenant` | Get current tenant |
| POST | `/api/v1/tenant/setup` | Create organization |
| GET | `/api/v1/tenant/plans` | List subscription plans |
| POST | `/api/v1/tenant/select-plan` | Select plan |
| GET | `/api/v1/tenant/check-slug` | Check slug availability |

### Workspace Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/workspaces` | List workspaces |
| POST | `/api/v1/workspaces` | Create workspace |
| GET | `/api/v1/workspaces/:id` | Get workspace |
| DELETE | `/api/v1/workspaces/:id` | Delete workspace |
| GET | `/api/v1/workspaces/:id/members` | List members |
| POST | `/api/v1/workspaces/:id/members` | Add member |

## React SDK Reference

### Hooks

#### `useSaasAuth()`

```tsx
const {
  user,              // Current user object
  isAuthenticated,   // Boolean
  isLoading,         // Loading state
  error,             // Auth error if any
  loginWithGoogle,   // () => Promise
  loginWithGithub,   // () => Promise
  loginWithEmail,    // (email, password) => Promise
  signup,            // (email, password, name) => Promise
  logout,            // () => void
} = useSaasAuth()
```

#### `useTenant()`

```tsx
const {
  tenant,            // Current tenant
  subscription,      // Active subscription
  plan,              // Current plan
  selectPlan,        // (planId) => Promise
  setupOrganization, // (name, slug) => Promise
  checkSlug,         // (slug) => Promise<boolean>
} = useTenant()
```

#### `useWorkspaces()`

```tsx
const {
  workspaces,          // List of workspaces
  currentWorkspace,    // Active workspace
  setCurrentWorkspace, // (id) => void
  createWorkspace,     // (name, slug) => Promise
} = useWorkspaces()
```

### Components

```tsx
import {
  ProtectedRoute,
  SocialLoginButtons,
} from '@saas-starter/react'

// Route protection
<ProtectedRoute
  fallback={<LoginPage />}
  requireTenant={true}
>
  <DashboardPage />
</ProtectedRoute>

// Social login buttons
<SocialLoginButtons
  providers={['google', 'github']}
  onSuccess={handleSuccess}
/>
```

## Authorization Model

The OpenFGA authorization model supports:

```
Platform
  └── admin: [user]

Tenant (Organization)
  ├── admin: [user]
  └── member: [user] or admin

Workspace (Project)
  ├── tenant: [tenant]
  ├── admin: [user] or tenant.can_manage
  ├── member: [user] or admin
  └── viewer: [user] or member
```

### Permissions

| Role | can_read | can_write | can_manage |
|------|----------|-----------|------------|
| viewer | ✓ | - | - |
| member | ✓ | ✓ | - |
| admin | ✓ | ✓ | ✓ |

## Environment Variables

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/saas

# Security
JWT_SECRET=your-secret-key
API_KEY_SECRET=your-api-key-secret

# Casdoor (IdP)
CASDOOR_ENDPOINT=http://localhost:8085
CASDOOR_CLIENT_ID=saas-client-id
CASDOOR_CLIENT_SECRET=saas-client-secret

# OpenFGA
OPENFGA_URL=http://localhost:8081
OPENFGA_STORE_ID=<store-id>

# Development
DEV_MODE=true  # Bypasses auth for local dev
```

## Development

### Dev Mode

Set `DEV_MODE=true` to bypass authentication during development. The AuthZ service will inject default user headers for all requests.

### Running Locally

```bash
# Start infrastructure only
docker compose up -d postgres openfga

# Run API locally
cd backend && go run ./cmd/api

# Run AuthZ locally
cd authz && go run ./cmd/authz
```

## Production Deployment

1. Set secure secrets in environment variables
2. Configure Casdoor with your OAuth providers
3. Create OpenFGA store and upload authorization model
4. Set `DEV_MODE=false`
5. Configure SSL/TLS on Traefik

## License

MIT
