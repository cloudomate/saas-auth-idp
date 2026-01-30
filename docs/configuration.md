# Configuration Guide

Complete reference for configuring the SaaS Auth IDP system.

## Environment Variables

All configuration is done through environment variables. Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
```

### Database Configuration

```bash
# PostgreSQL connection
POSTGRES_USER=saas
POSTGRES_PASSWORD=saas_password
POSTGRES_DB=saas_starter
DATABASE_URL=postgres://saas:saas_password@localhost:5432/saas_starter?sslmode=disable
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `POSTGRES_USER` | Yes | - | PostgreSQL username |
| `POSTGRES_PASSWORD` | Yes | - | PostgreSQL password |
| `POSTGRES_DB` | Yes | - | Database name |
| `DATABASE_URL` | Yes | - | Full connection string |

**Production Recommendations**:
- Use a managed PostgreSQL service (AWS RDS, GCP Cloud SQL, etc.)
- Enable SSL: `?sslmode=require`
- Use connection pooling for high traffic

### Security Configuration

```bash
# JWT signing secret (min 32 characters)
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# API key signing secret
API_KEY_SECRET=your-api-key-secret-change-in-production
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | Yes | - | Secret for signing JWT tokens |
| `API_KEY_SECRET` | Yes | - | Secret for API key validation |

**Security Best Practices**:
- Generate cryptographically secure secrets:
  ```bash
  openssl rand -base64 32
  ```
- Never commit secrets to version control
- Rotate secrets periodically
- Use different secrets for each environment

### Server URLs

```bash
# Backend API port
PORT=8000

# Public API URL (used for OAuth callbacks)
APP_URL=http://localhost:4455

# Frontend application URL (for CORS)
FRONTEND_URL=http://localhost:3000
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8000` | Backend API port |
| `APP_URL` | Yes | - | Public URL of the API gateway |
| `FRONTEND_URL` | Yes | - | Frontend URL for CORS headers |

### OAuth Providers

#### Google OAuth

```bash
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
```

**Setup**:
1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create OAuth 2.0 Client ID
3. Set authorized redirect URI: `{APP_URL}/api/v1/auth/social/callback`
4. Copy Client ID and Secret

#### GitHub OAuth

```bash
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
```

**Setup**:
1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Create new OAuth App
3. Set callback URL: `{APP_URL}/api/v1/auth/social/callback`
4. Copy Client ID and Secret

### OpenFGA Configuration

```bash
OPENFGA_URL=http://localhost:8081
OPENFGA_STORE_ID=01HXYZ...
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENFGA_URL` | Yes | - | OpenFGA service URL |
| `OPENFGA_STORE_ID` | Yes | - | OpenFGA store identifier |

**Initial Setup**:
```bash
# Create store
curl -X POST http://localhost:8081/stores \
  -H "Content-Type: application/json" \
  -d '{"name": "saas-starter"}'

# Note the store ID from response
# Upload authorization model
curl -X POST "http://localhost:8081/stores/{STORE_ID}/authorization-models" \
  -H "Content-Type: application/json" \
  -d @deploy/openfga/model.json
```

### Casdoor Configuration (Optional)

For enterprise SSO via Casdoor:

```bash
CASDOOR_ENDPOINT=http://localhost:8085
CASDOOR_CLIENT_ID=saas-client-id
CASDOOR_CLIENT_SECRET=saas-client-secret
```

### Development Mode

```bash
# Bypass authentication (local development only!)
DEV_MODE=true
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DEV_MODE` | No | `false` | Skip auth validation |

**Warning**: Never enable `DEV_MODE` in production!

When enabled:
- AuthZ service allows all requests
- Injects default headers:
  - `X-User-ID: dev-user-123`
  - `X-Tenant-ID: dev-tenant-456`
  - `X-Is-Platform-Admin: true`

## Docker Compose Configuration

### Service Ports

| Service | Internal Port | External Port | Purpose |
|---------|--------------|---------------|---------|
| API | 8000 | - | Backend (internal) |
| AuthZ | 8002 | - | ForwardAuth (internal) |
| Traefik | - | 4455 | API Gateway |
| Traefik | - | 8080 | Dashboard |
| PostgreSQL | 5432 | 5432 | Database |
| OpenFGA | 8081 | 8081 | Authorization |
| Casdoor | 8085 | 8085 | Identity Provider |

### Resource Limits

For production, set resource limits in `docker-compose.yml`:

```yaml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 128M
```

### Health Checks

```yaml
services:
  api:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

## Traefik Configuration

### Basic Configuration

Located at `deploy/traefik/traefik.yml`:

```yaml
entryPoints:
  web:
    address: ":4455"

api:
  dashboard: true
  insecure: true  # Disable in production

providers:
  docker:
    exposedByDefault: false
```

### ForwardAuth Middleware

```yaml
http:
  middlewares:
    auth:
      forwardAuth:
        address: "http://authz:8002/gate"
        trustForwardHeader: true
        authResponseHeaders:
          - "X-User-ID"
          - "X-Tenant-ID"
          - "X-Workspace-ID"
          - "X-Is-Platform-Admin"
```

### SSL/TLS (Production)

```yaml
entryPoints:
  websecure:
    address: ":443"

certificatesResolvers:
  letsencrypt:
    acme:
      email: admin@yourdomain.com
      storage: /letsencrypt/acme.json
      httpChallenge:
        entryPoint: web
```

## Hierarchy Configuration

Located at `deploy/hierarchy.json`:

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

See [Hierarchy Guide](./hierarchy.md) for customization options.

## Subscription Plans

Plans are seeded via database migration. To customize:

1. Edit the seed data in `backend/internal/models/seed.go`
2. Or insert directly into database:

```sql
INSERT INTO plans (id, tier, name, description, max_workspaces, max_users, monthly_price, annual_price, features, is_active)
VALUES (
  gen_random_uuid(),
  'custom',
  'Custom Plan',
  'Custom plan description',
  20,
  100,
  49.99,
  499.99,
  '["Feature 1", "Feature 2", "Feature 3"]',
  true
);
```

## CORS Configuration

CORS is configured in the backend middleware. Allowed origins:

```go
// backend/internal/api/middleware/cors.go
config := cors.Config{
    AllowOrigins:     []string{cfg.FrontendURL, "http://localhost:3000", "http://localhost:5173"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Authorization", "Content-Type", "X-Workspace-ID"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
}
```

To add custom origins, modify the middleware or set via environment:

```bash
CORS_ORIGINS=https://app.yourdomain.com,https://admin.yourdomain.com
```

## Logging Configuration

### Log Level

```bash
LOG_LEVEL=info  # debug, info, warn, error
```

### Log Format

```bash
LOG_FORMAT=json  # json, text
```

### Structured Logging

```go
// Example log output (JSON)
{
  "level": "info",
  "ts": "2024-01-15T10:30:00Z",
  "msg": "user_logged_in",
  "user_id": "123",
  "provider": "google"
}
```

## Production Deployment Checklist

### Security

- [ ] Set strong `JWT_SECRET` (32+ characters)
- [ ] Set strong `API_KEY_SECRET` (32+ characters)
- [ ] Disable `DEV_MODE`
- [ ] Enable HTTPS on Traefik
- [ ] Configure proper CORS origins
- [ ] Use managed database with SSL

### OAuth

- [ ] Configure Google OAuth credentials
- [ ] Configure GitHub OAuth credentials
- [ ] Set correct redirect URIs
- [ ] Test OAuth flow end-to-end

### Database

- [ ] Use managed PostgreSQL (RDS, Cloud SQL)
- [ ] Enable SSL connections
- [ ] Set up regular backups
- [ ] Configure connection pooling

### OpenFGA

- [ ] Create production store
- [ ] Upload authorization model
- [ ] Set `OPENFGA_STORE_ID`

### Monitoring

- [ ] Set up health check monitoring
- [ ] Configure log aggregation
- [ ] Set up error tracking (Sentry, etc.)
- [ ] Configure alerts

### Scaling

- [ ] Set resource limits
- [ ] Configure horizontal pod autoscaling
- [ ] Set up load balancer
- [ ] Configure database read replicas

## Example Production Configuration

```bash
# .env.production

# Database (managed PostgreSQL)
DATABASE_URL=postgres://user:pass@db.example.com:5432/saas?sslmode=require

# Security (strong secrets)
JWT_SECRET=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
API_KEY_SECRET=sk_live_abc123def456...

# URLs
APP_URL=https://api.yourdomain.com
FRONTEND_URL=https://app.yourdomain.com

# OAuth
GOOGLE_CLIENT_ID=123456789.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-xxx
GITHUB_CLIENT_ID=abc123
GITHUB_CLIENT_SECRET=ghp_xxx

# OpenFGA
OPENFGA_URL=https://openfga.yourdomain.com
OPENFGA_STORE_ID=01HXYZ...

# Production mode
DEV_MODE=false
LOG_LEVEL=info
LOG_FORMAT=json
```

## Troubleshooting Configuration

### Database Connection Failed

```
Error: failed to connect to database
```

1. Check `DATABASE_URL` format
2. Verify PostgreSQL is running
3. Check network connectivity
4. Verify credentials

### OAuth Redirect Mismatch

```
Error: redirect_uri_mismatch
```

1. Check `APP_URL` matches OAuth configuration
2. Verify redirect URI in provider console
3. Ensure protocol matches (http vs https)

### OpenFGA Connection Failed

```
Error: failed to connect to OpenFGA
```

1. Verify `OPENFGA_URL` is correct
2. Check OpenFGA service is running
3. Verify `OPENFGA_STORE_ID` exists

### CORS Errors

```
Error: blocked by CORS policy
```

1. Add frontend URL to CORS configuration
2. Check for trailing slashes
3. Verify protocol matches
