# Examples

This directory contains example applications demonstrating the SaaS Auth IDP system with ReBAC and ABAC authorization patterns.

## Quick Start

Start everything with a single command:

```bash
cd examples
docker compose up --build
```

This will start:
- **PostgreSQL** - Database for OpenFGA
- **OpenFGA** - Authorization engine (ReBAC)
- **Sample API** - Go backend demonstrating ReBAC/ABAC
- **React App** - Frontend with permission-aware UI

### Services & Ports

| Service | URL | Description |
|---------|-----|-------------|
| React App | http://localhost:3000 | Frontend application |
| Sample API | http://localhost:8001 | Go backend with ReBAC/ABAC |
| OpenFGA API | http://localhost:8081 | Authorization API |
| OpenFGA Playground | http://localhost:3001 | Visual authorization explorer |
| PostgreSQL | localhost:5433 | Database |

## What Gets Created

The setup automatically:

1. **Creates OpenFGA Store** with authorization model
2. **Loads Sample Tuples** for ReBAC demo:
   ```
   Workspace: workspace-1
   ├── user-1: admin
   ├── user-2: member
   └── user-3: viewer

   Documents:
   ├── doc-1: owner=user-1
   ├── doc-2: owner=user-2, editor=user-1
   └── doc-3: owner=user-3, viewers=[user-1, user-2]

   Projects:
   ├── proj-1: owner=user-1
   └── proj-2: owner=user-2, admin=user-1
   ```

3. **Seeds Sample Data** in the API for immediate testing

## Testing the Demo

### 1. Open React App

Go to http://localhost:3000 and navigate to:
- **Documents** - ReBAC demo (sharing, visibility)
- **Projects** - ABAC demo (environment policies)

### 2. Test with curl

**Test ReBAC (Documents):**
```bash
# User-1 (admin) can see all documents
curl http://localhost:8001/api/v1/documents \
  -H "X-User-ID: user-1" \
  -H "X-Workspace-ID: workspace-1"

# User-3 (viewer) has limited access
curl http://localhost:8001/api/v1/documents \
  -H "X-User-ID: user-3" \
  -H "X-Workspace-ID: workspace-1"

# Check specific permissions via OpenFGA
curl -X POST http://localhost:8001/api/v1/check-permission \
  -H "Content-Type: application/json" \
  -d '{"user": "user:user-2", "relation": "can_write", "object": "document:doc-1"}'
```

**Test ABAC (Projects):**
```bash
# Non-admin cannot deploy to production
curl -X POST http://localhost:8001/api/v1/projects/proj-1/deploy \
  -H "X-User-ID: user-2" \
  -H "X-Workspace-ID: workspace-1"
# Result: 403 - Only administrators can deploy to production

# Platform admin can deploy
curl -X POST http://localhost:8001/api/v1/projects/proj-1/deploy \
  -H "X-User-ID: user-1" \
  -H "X-Is-Platform-Admin: true" \
  -H "X-Workspace-ID: workspace-1"
# Result: 200 - Deployment initiated
```

### 3. Explore OpenFGA Playground

Visit http://localhost:3001 to visually explore:
- Authorization model
- Relationship tuples
- Permission queries

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Docker Compose                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐   │
│  │  React App  │────▶│ Sample API  │────▶│  OpenFGA    │   │
│  │  (Port 3000)│     │ (Port 8001) │     │ (Port 8081) │   │
│  └─────────────┘     └─────────────┘     └─────────────┘   │
│                             │                    │          │
│                             │                    │          │
│                             ▼                    ▼          │
│                      ┌─────────────┐     ┌─────────────┐   │
│                      │  In-Memory  │     │ PostgreSQL  │   │
│                      │    Store    │     │ (Port 5433) │   │
│                      │   (ABAC)    │     │   (ReBAC)   │   │
│                      └─────────────┘     └─────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘

Authorization Flow:
1. Request arrives at Sample API
2. For ReBAC: API queries OpenFGA for relationship checks
3. For ABAC: API evaluates policies based on attributes
4. Combined decision determines access
```

## Authorization Patterns

### ReBAC (Documents)

Access based on **relationships**:

```
User ──[relationship]──▶ Document

Relationships: owner, editor, viewer
Permissions derived:
  - owner  → can_read, can_write, can_delete, can_share
  - editor → can_read, can_write
  - viewer → can_read
```

### ABAC (Projects)

Access based on **attributes**:

```
Policies:
  - Production projects: admin-only for write/deploy
  - Archived projects: read-only
  - Paused projects: no deployment
  - Delete: owner or admin only
```

## Development Mode

For hot-reloading during development:

```bash
docker compose -f docker-compose.dev.yml up --build
```

## Running Without Docker

### Sample API
```bash
cd sample-api
go run main.go
```

### React App
```bash
cd react-app
npm install
npm run dev
```

## Cleanup

```bash
# Stop and remove containers
docker compose down

# Remove volumes (database data)
docker compose down -v
```

## Files Structure

```
examples/
├── docker-compose.yml          # Main compose file
├── docker-compose.dev.yml      # Development compose
├── deploy/
│   ├── init.sql               # PostgreSQL init
│   ├── model.json             # OpenFGA authorization model
│   └── setup-openfga.sh       # OpenFGA setup script
├── react-app/                  # React frontend
│   ├── Dockerfile
│   ├── nginx.conf
│   └── src/
└── sample-api/                 # Go backend
    ├── Dockerfile
    ├── main.go
    └── internal/
        ├── handlers/          # ReBAC & ABAC handlers
        ├── authz/             # OpenFGA client
        └── store/             # In-memory data store
```
