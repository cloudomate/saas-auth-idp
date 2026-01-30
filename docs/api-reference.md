# API Reference

Complete REST API documentation for the SaaS Auth IDP backend.

## Base URL

All API requests should be made to:
- **Development**: `http://localhost:4455`
- **Production**: `https://api.yourdomain.com`

## Authentication

Most endpoints require authentication via JWT token or API key.

### JWT Token

```bash
curl -H "Authorization: Bearer <jwt_token>" \
     http://localhost:4455/api/v1/workspaces
```

### API Key

API keys are workspace-scoped and require the workspace ID header:

```bash
curl -H "Authorization: Bearer sk-<key_id>-<secret>" \
     -H "X-Workspace-ID: <workspace_uuid>" \
     http://localhost:4455/api/v1/workspaces
```

## Response Format

### Success Response

```json
{
  "data": { ... },
  "message": "Success message (optional)"
}
```

### Error Response

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": { ... }
}
```

### Common HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found |
| 409 | Conflict - Resource already exists |
| 500 | Internal Server Error |

---

## Authentication Endpoints

### Get OAuth URL

Initiates OAuth flow with a provider.

```
GET /api/v1/auth/social/:provider/login
```

**Parameters**:
- `provider` (path): `google` or `github`
- `redirect_uri` (query, optional): Frontend redirect after auth
- `flow` (query, optional): `login` or `signup` (default: `login`)
- `plan` (query, optional): Pre-select plan tier

**Response**:
```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/auth?...",
  "state": "abc123..."
}
```

**Example**:
```bash
curl "http://localhost:4455/api/v1/auth/social/google/login?plan=basic"
```

### OAuth Callback

Handles OAuth callback after provider authorization.

```
POST /api/v1/auth/social/callback
```

**Request Body**:
```json
{
  "code": "authorization_code_from_provider",
  "state": "state_token_from_initiate"
}
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "picture": "https://...",
    "auth_provider": "google",
    "email_verified": true,
    "is_tenant_admin": false,
    "created_at": "2024-01-15T10:30:00Z"
  },
  "needs_tenant_setup": true,
  "flow": "login"
}
```

### Register (Email/Password)

Create a new account with email and password.

```
POST /api/v1/auth/register
```

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "John Doe",
  "plan": "basic"
}
```

**Response** (201 Created):
```json
{
  "message": "Account created. Please check your email to verify your account.",
  "verify_token": "abc123..."
}
```

**Errors**:
- `email_exists`: Account with email already exists

### Verify Email

Verify email address with token.

```
POST /api/v1/auth/verify-email
```

**Request Body**:
```json
{
  "token": "verification_token"
}
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { ... },
  "needs_tenant_setup": true
}
```

**Errors**:
- `invalid_token`: Token not found
- `token_expired`: Verification token expired (24h)

### Login (Email/Password)

Authenticate with email and password.

```
POST /api/v1/auth/login
```

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "auth_provider": "local",
    "email_verified": true,
    "is_tenant_admin": true,
    "tenant_id": "660e8400-e29b-41d4-a716-446655440001",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "needs_tenant_setup": false
}
```

**Errors**:
- `invalid_credentials`: Wrong email or password
- `email_not_verified`: Email needs verification

### Forgot Password

Request password reset email.

```
POST /api/v1/auth/forgot-password
```

**Request Body**:
```json
{
  "email": "user@example.com"
}
```

**Response**:
```json
{
  "message": "If an account exists, a reset link has been sent.",
  "reset_token": "xyz789..."
}
```

Note: Response is always successful to prevent email enumeration.

### Reset Password

Reset password with token.

```
POST /api/v1/auth/reset-password
```

**Request Body**:
```json
{
  "token": "reset_token",
  "password": "newSecurePassword123"
}
```

**Response**:
```json
{
  "message": "Password reset successful"
}
```

**Errors**:
- `invalid_token`: Token not found
- `token_expired`: Reset token expired (1h)

### Get Current User

Get authenticated user's profile.

```
GET /api/v1/auth/me
```

**Headers**: `Authorization: Bearer <token>`

**Response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "picture": "https://...",
  "auth_provider": "google",
  "email_verified": true,
  "is_tenant_admin": true,
  "tenant_id": "660e8400-e29b-41d4-a716-446655440001",
  "selected_plan": "advanced",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

## Tenant Endpoints

### Get Current Tenant

Get the authenticated user's tenant.

```
GET /api/v1/tenant
```

**Headers**: `Authorization: Bearer <token>`

**Response**:
```json
{
  "tenant": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "level": "tenant",
    "slug": "acme-inc",
    "display_name": "Acme Inc",
    "root_id": "660e8400-e29b-41d4-a716-446655440001",
    "depth": 0,
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z"
  },
  "subscription": {
    "id": "770e8400-e29b-41d4-a716-446655440002",
    "tenant_id": "660e8400-e29b-41d4-a716-446655440001",
    "plan_id": "880e8400-e29b-41d4-a716-446655440003",
    "status": "active",
    "current_period_start": "2024-01-15T00:00:00Z",
    "current_period_end": "2024-02-15T00:00:00Z",
    "plan": {
      "id": "880e8400-e29b-41d4-a716-446655440003",
      "tier": "advanced",
      "name": "Advanced",
      "max_workspaces": 10,
      "max_users": 50
    }
  }
}
```

### List Plans

Get available subscription plans (public endpoint).

```
GET /api/v1/tenant/plans
```

**Response**:
```json
{
  "plans": [
    {
      "id": "880e8400-e29b-41d4-a716-446655440001",
      "tier": "basic",
      "name": "Basic",
      "description": "For individuals and small teams",
      "max_workspaces": 3,
      "max_users": 5,
      "monthly_price": 0,
      "annual_price": 0,
      "allows_on_prem": false,
      "features": ["3 workspaces", "5 team members", "Basic support"],
      "is_active": true
    },
    {
      "id": "880e8400-e29b-41d4-a716-446655440002",
      "tier": "advanced",
      "name": "Advanced",
      "description": "For growing teams",
      "max_workspaces": 10,
      "max_users": 50,
      "monthly_price": 29,
      "annual_price": 290,
      "allows_on_prem": false,
      "features": ["10 workspaces", "50 team members", "Priority support"],
      "is_active": true
    },
    {
      "id": "880e8400-e29b-41d4-a716-446655440003",
      "tier": "enterprise",
      "name": "Enterprise",
      "description": "For large organizations",
      "max_workspaces": -1,
      "max_users": -1,
      "monthly_price": 99,
      "annual_price": 990,
      "allows_on_prem": true,
      "features": ["Unlimited workspaces", "Unlimited users", "SSO", "Dedicated support"],
      "is_active": true
    }
  ]
}
```

### Select Plan

Select a subscription plan. For "basic" tier, automatically creates tenant.

```
POST /api/v1/tenant/select-plan
```

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "plan": "basic"
}
```

**Response** (Basic plan - auto-creates tenant):
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "tenant": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "slug": "john-doe",
    "display_name": "John Doe's Organization",
    ...
  },
  "workspace": {
    "id": "990e8400-e29b-41d4-a716-446655440001",
    "slug": "default",
    "display_name": "Default Workspace",
    ...
  }
}
```

**Response** (Advanced/Enterprise - requires setup):
```json
{
  "message": "Plan selected. Please complete organization setup.",
  "selected_plan": "advanced"
}
```

### Setup Organization

Create organization after selecting non-basic plan.

```
POST /api/v1/tenant/setup
```

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "name": "Acme Inc",
  "slug": "acme-inc"
}
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "tenant": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "slug": "acme-inc",
    "display_name": "Acme Inc",
    ...
  },
  "workspace": {
    "id": "990e8400-e29b-41d4-a716-446655440001",
    "slug": "default",
    "display_name": "Default Workspace",
    ...
  }
}
```

**Errors**:
- `slug_taken`: Slug already in use

### Check Slug Availability

Check if a slug is available.

```
GET /api/v1/tenant/check-slug?slug=acme-inc
```

**Response**:
```json
{
  "available": true,
  "slug": "acme-inc"
}
```

---

## Workspace Endpoints

### List Workspaces

Get all workspaces in the tenant.

```
GET /api/v1/workspaces
```

**Headers**: `Authorization: Bearer <token>`

**Response**:
```json
{
  "workspaces": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440001",
      "level": "workspace",
      "slug": "default",
      "display_name": "Default Workspace",
      "parent_id": "660e8400-e29b-41d4-a716-446655440001",
      "root_id": "660e8400-e29b-41d4-a716-446655440001",
      "depth": 1,
      "is_active": true,
      "is_default": true,
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "990e8400-e29b-41d4-a716-446655440002",
      "level": "workspace",
      "slug": "engineering",
      "display_name": "Engineering",
      "parent_id": "660e8400-e29b-41d4-a716-446655440001",
      "root_id": "660e8400-e29b-41d4-a716-446655440001",
      "depth": 1,
      "is_active": true,
      "is_default": false,
      "created_at": "2024-01-16T14:00:00Z"
    }
  ]
}
```

### Create Workspace

Create a new workspace.

```
POST /api/v1/workspaces
```

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "name": "Engineering",
  "slug": "engineering"
}
```

**Response** (201 Created):
```json
{
  "workspace": {
    "id": "990e8400-e29b-41d4-a716-446655440002",
    "level": "workspace",
    "slug": "engineering",
    "display_name": "Engineering",
    "parent_id": "660e8400-e29b-41d4-a716-446655440001",
    "root_id": "660e8400-e29b-41d4-a716-446655440001",
    "depth": 1,
    "is_active": true,
    "created_at": "2024-01-16T14:00:00Z"
  }
}
```

**Errors**:
- `limit_exceeded`: Workspace limit reached for plan

### Get Workspace

Get a specific workspace by ID or slug.

```
GET /api/v1/workspaces/:id
```

**Parameters**:
- `id`: UUID or slug

**Headers**: `Authorization: Bearer <token>`

**Response**:
```json
{
  "workspace": {
    "id": "990e8400-e29b-41d4-a716-446655440001",
    "level": "workspace",
    "slug": "default",
    "display_name": "Default Workspace",
    ...
  }
}
```

### Delete Workspace

Delete a workspace.

```
DELETE /api/v1/workspaces/:id
```

**Headers**: `Authorization: Bearer <token>`

**Response**:
```json
{
  "message": "Workspace deleted successfully"
}
```

**Errors**:
- `cannot_delete_default`: Cannot delete default workspace

### List Workspace Members

Get members of a workspace.

```
GET /api/v1/workspaces/:id/members
```

**Headers**: `Authorization: Bearer <token>`

**Response**:
```json
{
  "members": [
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "admin@example.com",
      "name": "John Doe",
      "picture": "https://...",
      "role": "admin",
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440001",
      "email": "member@example.com",
      "name": "Jane Smith",
      "picture": "https://...",
      "role": "member",
      "created_at": "2024-01-16T14:00:00Z"
    }
  ]
}
```

### Add Workspace Member

Add a member to a workspace.

```
POST /api/v1/workspaces/:id/members
```

**Headers**: `Authorization: Bearer <token>`

**Request Body**:
```json
{
  "email": "newmember@example.com",
  "role": "member"
}
```

**Response**:
```json
{
  "message": "Member added successfully",
  "membership": {
    "user_id": "550e8400-e29b-41d4-a716-446655440002",
    "workspace_id": "990e8400-e29b-41d4-a716-446655440001",
    "role": "member",
    "created_at": "2024-01-17T09:00:00Z"
  }
}
```

**Roles**:
- `admin`: Full control
- `member`: Read/write access
- `viewer`: Read-only access

---

## Hierarchy Endpoint

### Get Hierarchy Configuration

Get the organization hierarchy configuration.

```
GET /api/v1/hierarchy
```

**Headers**: `Authorization: Bearer <token>`

**Response**:
```json
{
  "root_level": "tenant",
  "leaf_level": "workspace",
  "depth": 2,
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

---

## Health Check

### Health

Check service health.

```
GET /api/v1/health
```

**Response**:
```json
{
  "status": "ok"
}
```

---

## Error Codes Reference

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid_request` | 400 | Missing or invalid parameters |
| `invalid_credentials` | 401 | Wrong email or password |
| `unauthorized` | 401 | Missing or invalid token |
| `forbidden` | 403 | Insufficient permissions |
| `email_not_verified` | 403 | Email needs verification |
| `not_found` | 404 | Resource not found |
| `email_exists` | 409 | Email already registered |
| `slug_taken` | 409 | Slug already in use |
| `limit_exceeded` | 409 | Plan limit reached |
| `invalid_token` | 400 | Invalid verification/reset token |
| `token_expired` | 400 | Token has expired |
| `state_expired` | 400 | OAuth state expired |
| `provider_not_configured` | 400 | OAuth provider not set up |
| `internal_error` | 500 | Server error |
