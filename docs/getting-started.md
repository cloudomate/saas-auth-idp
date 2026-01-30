# Getting Started

This guide will help you set up the SaaS Auth IDP system and integrate it with your application.

## Prerequisites

- Docker and Docker Compose
- Node.js 18+ (for React SDK development)
- Go 1.21+ (optional, for local backend development)

## Quick Start

### 1. Clone and Configure

```bash
git clone https://github.com/yourusername/saas-auth-idp.git
cd saas-auth-idp

# Copy environment template
cp .env.example .env
```

### 2. Configure OAuth Providers (Optional)

Edit `.env` to add your OAuth credentials:

```bash
# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# GitHub OAuth
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
```

To get OAuth credentials:
- **Google**: [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
- **GitHub**: [GitHub Developer Settings](https://github.com/settings/developers)

Set the redirect URI to: `http://localhost:4455/api/v1/auth/social/callback`

### 3. Start the Services

```bash
docker compose up -d
```

This starts:
| Service | Port | URL |
|---------|------|-----|
| API Gateway | 4455 | http://localhost:4455 |
| Traefik Dashboard | 8080 | http://localhost:8080 |
| PostgreSQL | 5432 | - |
| OpenFGA | 8081 | http://localhost:8081 |

### 4. Verify Installation

```bash
# Check health endpoint
curl http://localhost:4455/api/v1/health

# Expected: {"status":"ok"}
```

## Frontend Integration

### Install the React SDK

```bash
npm install @saas-starter/react
# or
yarn add @saas-starter/react
```

### Basic Setup

```tsx
// App.tsx
import { SaasAuthProvider } from '@saas-starter/react'
import { AuthenticatedApp } from './AuthenticatedApp'

function App() {
  return (
    <SaasAuthProvider
      apiUrl="http://localhost:4455"
      storagePrefix="myapp"
    >
      <AuthenticatedApp />
    </SaasAuthProvider>
  )
}

export default App
```

### Authentication Flow

```tsx
// AuthenticatedApp.tsx
import { useSaasAuth, useTenant, useWorkspaces } from '@saas-starter/react'

export function AuthenticatedApp() {
  const { user, isAuthenticated, isLoading, loginWithGoogle, logout } = useSaasAuth()
  const { tenant, needsTenantSetup } = useTenant()
  const { workspaces, currentWorkspace } = useWorkspaces()

  // Loading state
  if (isLoading) {
    return <div>Loading...</div>
  }

  // Not authenticated - show login
  if (!isAuthenticated) {
    return (
      <div>
        <h1>Welcome</h1>
        <button onClick={() => loginWithGoogle()}>
          Sign in with Google
        </button>
      </div>
    )
  }

  // Needs organization setup
  if (needsTenantSetup) {
    return <OrganizationSetup />
  }

  // Authenticated - show dashboard
  return (
    <div>
      <header>
        <span>Welcome, {user.name}!</span>
        <span>Organization: {tenant?.display_name}</span>
        <button onClick={logout}>Logout</button>
      </header>
      <main>
        <h2>Your Workspaces</h2>
        <ul>
          {workspaces.map(ws => (
            <li key={ws.id}>{ws.display_name}</li>
          ))}
        </ul>
      </main>
    </div>
  )
}
```

### Organization Setup Component

```tsx
// OrganizationSetup.tsx
import { useState } from 'react'
import { useTenant } from '@saas-starter/react'

export function OrganizationSetup() {
  const { plans, selectPlan, setupOrganization, checkSlug } = useTenant()
  const [step, setStep] = useState<'plan' | 'org'>('plan')
  const [selectedPlan, setSelectedPlan] = useState('')
  const [orgName, setOrgName] = useState('')
  const [slug, setSlug] = useState('')
  const [slugAvailable, setSlugAvailable] = useState(true)

  const handlePlanSelect = async (planTier: string) => {
    setSelectedPlan(planTier)
    await selectPlan(planTier)

    // Basic plan auto-creates org, others need setup
    if (planTier !== 'basic') {
      setStep('org')
    }
  }

  const handleSlugChange = async (value: string) => {
    setSlug(value)
    if (value.length >= 3) {
      const available = await checkSlug(value)
      setSlugAvailable(available)
    }
  }

  const handleOrgSetup = async () => {
    await setupOrganization(orgName, slug)
  }

  if (step === 'plan') {
    return (
      <div>
        <h1>Choose Your Plan</h1>
        <div className="plans">
          {plans.map(plan => (
            <div key={plan.id} className="plan-card">
              <h3>{plan.name}</h3>
              <p>{plan.description}</p>
              <p>${plan.monthly_price}/month</p>
              <ul>
                {plan.features.map((f, i) => <li key={i}>{f}</li>)}
              </ul>
              <button onClick={() => handlePlanSelect(plan.tier)}>
                Select {plan.name}
              </button>
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div>
      <h1>Set Up Your Organization</h1>
      <form onSubmit={(e) => { e.preventDefault(); handleOrgSetup() }}>
        <div>
          <label>Organization Name</label>
          <input
            value={orgName}
            onChange={(e) => setOrgName(e.target.value)}
            placeholder="Acme Inc"
          />
        </div>
        <div>
          <label>URL Slug</label>
          <input
            value={slug}
            onChange={(e) => handleSlugChange(e.target.value)}
            placeholder="acme"
          />
          {!slugAvailable && <span className="error">Slug not available</span>}
        </div>
        <button type="submit" disabled={!slugAvailable || !orgName}>
          Create Organization
        </button>
      </form>
    </div>
  )
}
```

## Development Mode

For local development, you can bypass authentication:

```bash
# In .env
DEV_MODE=true
```

This injects default user headers for all requests, allowing you to develop without OAuth setup.

## Next Steps

- [Architecture Overview](./architecture.md) - Understand the system design
- [React SDK Reference](./react-sdk.md) - Complete SDK documentation
- [API Reference](./api-reference.md) - REST API endpoints
- [Authentication Guide](./authentication.md) - Auth flows in detail
- [Authorization Guide](./authorization.md) - Permissions and RBAC
- [Hierarchy Configuration](./hierarchy.md) - Custom organizational structures
