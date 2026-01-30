# React SDK Reference

Complete reference for the `@saas-starter/react` SDK.

## Installation

```bash
npm install @saas-starter/react
# or
yarn add @saas-starter/react
```

## Setup

### SaasAuthProvider

Wrap your application with the provider:

```tsx
import { SaasAuthProvider } from '@saas-starter/react'

function App() {
  return (
    <SaasAuthProvider
      apiUrl="http://localhost:4455"
      storagePrefix="myapp"
      onAuthStateChange={(user) => {
        console.log('Auth state changed:', user)
      }}
    >
      <YourApp />
    </SaasAuthProvider>
  )
}
```

**Props**:

| Prop | Type | Required | Description |
|------|------|----------|-------------|
| `apiUrl` | `string` | Yes | Backend API URL |
| `storagePrefix` | `string` | No | LocalStorage key prefix (default: `saas`) |
| `onAuthStateChange` | `(user: User \| null) => void` | No | Callback on auth changes |

## Hooks

### useSaasAuth

Primary hook for authentication operations.

```tsx
import { useSaasAuth } from '@saas-starter/react'

function LoginPage() {
  const {
    // State
    user,
    isAuthenticated,
    isLoading,
    error,
    needsTenantSetup,

    // OAuth
    loginWithGoogle,
    loginWithGithub,
    handleOAuthCallback,

    // Email/Password
    loginWithEmail,
    signup,
    verifyEmail,
    forgotPassword,
    resetPassword,

    // Session
    logout,
    refreshUser,
  } = useSaasAuth()

  // ...
}
```

#### State Properties

| Property | Type | Description |
|----------|------|-------------|
| `user` | `User \| null` | Current authenticated user |
| `isAuthenticated` | `boolean` | Whether user is logged in |
| `isLoading` | `boolean` | Loading state during auth operations |
| `error` | `string \| null` | Error message if any |
| `needsTenantSetup` | `boolean` | Whether user needs to create organization |

#### OAuth Methods

```tsx
// Redirect to Google OAuth
await loginWithGoogle(redirectUri?: string)

// Redirect to GitHub OAuth
await loginWithGithub(redirectUri?: string)

// Handle OAuth callback (call after redirect)
await handleOAuthCallback(code: string, state: string)
```

#### Email/Password Methods

```tsx
// Register new user
await signup(email: string, password: string, name?: string)

// Login with email/password
await loginWithEmail(email: string, password: string)

// Verify email with token
await verifyEmail(token: string)

// Request password reset
await forgotPassword(email: string)

// Reset password with token
await resetPassword(token: string, newPassword: string)
```

#### Session Methods

```tsx
// Logout and clear session
logout()

// Refresh user data from server
await refreshUser()
```

### useTenant

Hook for tenant/organization and subscription management.

```tsx
import { useTenant } from '@saas-starter/react'

function OrganizationSettings() {
  const {
    // State
    tenant,
    subscription,
    plan,
    plans,
    isLoading,
    error,

    // Actions
    selectPlan,
    setupOrganization,
    checkSlug,
    refreshTenant,
  } = useTenant()

  // ...
}
```

#### State Properties

| Property | Type | Description |
|----------|------|-------------|
| `tenant` | `Container \| null` | Current organization |
| `subscription` | `Subscription \| null` | Active subscription |
| `plan` | `Plan \| null` | Current plan details |
| `plans` | `Plan[]` | Available plans |
| `isLoading` | `boolean` | Loading state |
| `error` | `{ error: string; message: string } \| null` | Error if any |

#### Methods

```tsx
// Select a subscription plan
// For 'basic', auto-creates tenant
// For 'advanced'/'enterprise', requires setupOrganization()
await selectPlan(planTier: 'basic' | 'advanced' | 'enterprise')

// Create organization (after selecting non-basic plan)
await setupOrganization(name: string, slug: string)

// Check if slug is available
const isAvailable = await checkSlug(slug: string)

// Refresh tenant data
await refreshTenant()
```

### useWorkspaces

Hook for workspace management.

```tsx
import { useWorkspaces } from '@saas-starter/react'

function WorkspaceSelector() {
  const {
    // State
    workspaces,
    currentWorkspace,
    isLoading,
    error,

    // Actions
    setCurrentWorkspace,
    createWorkspace,
    refreshWorkspaces,
  } = useWorkspaces()

  return (
    <select
      value={currentWorkspace?.id}
      onChange={(e) => setCurrentWorkspace(e.target.value)}
    >
      {workspaces.map(ws => (
        <option key={ws.id} value={ws.id}>
          {ws.display_name}
        </option>
      ))}
    </select>
  )
}
```

#### State Properties

| Property | Type | Description |
|----------|------|-------------|
| `workspaces` | `Container[]` | List of workspaces |
| `currentWorkspace` | `Container \| null` | Selected workspace |
| `isLoading` | `boolean` | Loading state |
| `error` | `{ error: string; message: string } \| null` | Error if any |

#### Methods

```tsx
// Set active workspace
setCurrentWorkspace(workspaceId: string)

// Create new workspace
const workspace = await createWorkspace(name: string, slug: string)

// Refresh workspace list
await refreshWorkspaces()
```

### useHierarchy

Advanced hook for pluggable hierarchy operations. Use this for custom organizational structures.

```tsx
import { useHierarchy } from '@saas-starter/react'

function HierarchyNavigator() {
  const {
    // State
    config,
    isLoading,
    error,
    currentContainers,

    // Convenience getters
    rootLevel,
    leafLevel,
    currentRoot,
    currentLeaf,
    getLevel,

    // Actions
    fetchConfig,
    listContainers,
    createContainer,
    getContainer,
    deleteContainer,
    listMembers,
    addMember,
    setCurrentContainer,
  } = useHierarchy()

  // Display hierarchy levels
  return (
    <div>
      {config?.levels.map(level => (
        <div key={level.name}>
          <h3>{level.display_name}</h3>
          <p>Roles: {level.roles.join(', ')}</p>
        </div>
      ))}
    </div>
  )
}
```

#### State Properties

| Property | Type | Description |
|----------|------|-------------|
| `config` | `HierarchyConfig \| null` | Hierarchy configuration |
| `isLoading` | `boolean` | Loading state |
| `error` | `string \| null` | Error if any |
| `currentContainers` | `Record<string, Container \| null>` | Current selection per level |
| `rootLevel` | `HierarchyLevel \| null` | Root level config (tenant) |
| `leafLevel` | `HierarchyLevel \| null` | Leaf level config (workspace) |
| `currentRoot` | `Container \| null` | Current root container |
| `currentLeaf` | `Container \| null` | Current leaf container |

#### Methods

```tsx
// Get level configuration by name
const level = getLevel(levelName: string)

// Fetch hierarchy config from server
await fetchConfig()

// List containers at a level
const containers = await listContainers(level: string, parentId?: string)

// Create container at a level
const container = await createContainer(
  level: string,
  name: string,
  slug?: string,
  parentId?: string
)

// Get specific container
const container = await getContainer(level: string, idOrSlug: string)

// Delete container
await deleteContainer(level: string, id: string)

// List members of a container
const members = await listMembers(level: string, containerId: string)

// Add member to container
await addMember(
  level: string,
  containerId: string,
  email: string,
  role?: string
)

// Set current container for a level
setCurrentContainer(level: string, container: Container | null)
```

### useContainers

Generic hook for working with a specific hierarchy level.

```tsx
import { useContainers } from '@saas-starter/react'

function ProjectList() {
  const {
    containers,
    currentContainer,
    isLoading,
    error,
    level,
    setCurrentContainer,
    create,
    delete: deleteContainer,
    listMembers,
    addMember,
  } = useContainers('project')

  return (
    <div>
      <h2>{level.display_name}s</h2>
      {containers.map(c => (
        <div key={c.id}>{c.display_name}</div>
      ))}
    </div>
  )
}
```

## Components

### ProtectedRoute

Route guard component that handles authentication and tenant requirements.

```tsx
import { ProtectedRoute } from '@saas-starter/react'

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/dashboard"
        element={
          <ProtectedRoute
            fallback={<Navigate to="/login" />}
            loadingComponent={<Spinner />}
            requireTenant={true}
            tenantSetupComponent={<OrganizationSetup />}
          >
            <DashboardPage />
          </ProtectedRoute>
        }
      />
    </Routes>
  )
}
```

**Props**:

| Prop | Type | Required | Description |
|------|------|----------|-------------|
| `children` | `ReactNode` | Yes | Protected content |
| `fallback` | `ReactNode` | No | Shown when not authenticated |
| `loadingComponent` | `ReactNode` | No | Shown during loading |
| `requireTenant` | `boolean` | No | Require tenant setup |
| `tenantSetupComponent` | `ReactNode` | No | Shown when tenant setup needed |

### SocialLoginButtons

Pre-built OAuth login buttons.

```tsx
import { SocialLoginButtons } from '@saas-starter/react'

function LoginPage() {
  return (
    <SocialLoginButtons
      providers={['google', 'github']}
      onSuccess={(user) => {
        console.log('Logged in:', user)
      }}
      onError={(error) => {
        console.error('Login failed:', error)
      }}
      buttonClassName="social-btn"
      containerClassName="social-btns"
    />
  )
}
```

**Props**:

| Prop | Type | Required | Description |
|------|------|----------|-------------|
| `providers` | `('google' \| 'github')[]` | Yes | Providers to show |
| `onSuccess` | `(user: User) => void` | No | Success callback |
| `onError` | `(error: string) => void` | No | Error callback |
| `buttonClassName` | `string` | No | Button CSS class |
| `containerClassName` | `string` | No | Container CSS class |

## Types

### User

```typescript
interface User {
  id: string
  email: string
  name: string
  picture?: string
  auth_provider: 'google' | 'github' | 'local'
  email_verified: boolean
  is_platform_admin?: boolean
  is_root_admin?: boolean
  admin_of_root_id?: string
  selected_plan?: string
  tenantId?: string
  created_at: string
}
```

### Container (Tenant/Workspace)

```typescript
interface Container {
  id: string
  level: string
  slug: string
  display_name: string
  parent_id?: string
  root_id: string
  depth: number
  is_active: boolean
  created_at: string
  metadata?: Record<string, unknown>
}

// Aliases for backward compatibility
type Tenant = Container
type Workspace = Container
```

### Plan

```typescript
interface Plan {
  id: string
  tier: 'basic' | 'advanced' | 'enterprise'
  name: string
  description: string
  max_workspaces: number
  max_users: number
  monthly_price: number
  annual_price: number
  allows_on_prem: boolean
  features: string[]
  is_active: boolean
}
```

### Subscription

```typescript
interface Subscription {
  id: string
  tenant_id: string
  plan_id: string
  status: 'active' | 'cancelled' | 'past_due' | 'trialing'
  current_period_start: string
  current_period_end: string
  plan?: Plan
}
```

### Membership

```typescript
interface Membership {
  user_id: string
  email: string
  name: string
  picture?: string
  role: string
  created_at: string
}
```

### HierarchyConfig

```typescript
interface HierarchyConfig {
  root_level: string
  leaf_level: string
  depth: number
  levels: HierarchyLevel[]
}

interface HierarchyLevel {
  name: string
  display_name: string
  plural: string
  url_path: string
  roles: string[]
  is_root: boolean
}
```

### AuthError

```typescript
interface AuthError {
  code: string
  message: string
}

interface ApiError {
  error: string
  message: string
  details?: unknown
}
```

## Utilities

### SaasApiClient

Low-level API client for custom requests.

```tsx
import { SaasApiClient } from '@saas-starter/react'

const client = new SaasApiClient({
  apiUrl: 'http://localhost:4455',
  storagePrefix: 'myapp',
})

// Make authenticated request
const data = await client.get('/api/v1/custom-endpoint')
const result = await client.post('/api/v1/custom', { foo: 'bar' })
```

### Token Utilities

```tsx
import { decodeJwt, isTokenExpired } from '@saas-starter/react'

// Decode JWT without verification
const claims = decodeJwt(token)

// Check if token is expired
const expired = isTokenExpired(token)
```

### Storage

```tsx
import { createStorage } from '@saas-starter/react'

const storage = createStorage('myapp')

storage.setItem('key', 'value')
const value = storage.getItem('key')
storage.removeItem('key')
```

## Error Handling

All async methods can throw errors. Handle them appropriately:

```tsx
import { useSaasAuth } from '@saas-starter/react'

function LoginForm() {
  const { loginWithEmail, error } = useSaasAuth()
  const [localError, setLocalError] = useState('')

  const handleLogin = async (email, password) => {
    try {
      await loginWithEmail(email, password)
    } catch (err) {
      // err is an ApiError
      setLocalError(err.message)
    }
  }

  return (
    <form>
      {(error || localError) && (
        <div className="error">{error || localError}</div>
      )}
      {/* form fields */}
    </form>
  )
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `invalid_credentials` | Wrong email or password |
| `email_exists` | Email already registered |
| `email_not_verified` | Email needs verification |
| `invalid_token` | Invalid or expired token |
| `token_expired` | Verification/reset token expired |
| `provider_not_configured` | OAuth provider not set up |
| `slug_taken` | Organization slug not available |

## Complete Example

```tsx
import {
  SaasAuthProvider,
  useSaasAuth,
  useTenant,
  useWorkspaces,
  ProtectedRoute,
} from '@saas-starter/react'

function App() {
  return (
    <SaasAuthProvider apiUrl="http://localhost:4455">
      <Router>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/*"
            element={
              <ProtectedRoute requireTenant>
                <AuthenticatedRoutes />
              </ProtectedRoute>
            }
          />
        </Routes>
      </Router>
    </SaasAuthProvider>
  )
}

function LoginPage() {
  const { loginWithGoogle, loginWithEmail, isLoading, error } = useSaasAuth()

  if (isLoading) return <div>Loading...</div>

  return (
    <div>
      <h1>Login</h1>
      {error && <div className="error">{error}</div>}
      <button onClick={() => loginWithGoogle()}>
        Continue with Google
      </button>
      <EmailLoginForm onSubmit={loginWithEmail} />
    </div>
  )
}

function AuthenticatedRoutes() {
  const { user, logout } = useSaasAuth()
  const { tenant } = useTenant()
  const { workspaces, currentWorkspace, setCurrentWorkspace } = useWorkspaces()

  return (
    <div>
      <nav>
        <span>{user.name} @ {tenant?.display_name}</span>
        <select
          value={currentWorkspace?.id}
          onChange={(e) => setCurrentWorkspace(e.target.value)}
        >
          {workspaces.map(ws => (
            <option key={ws.id} value={ws.id}>{ws.display_name}</option>
          ))}
        </select>
        <button onClick={logout}>Logout</button>
      </nav>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </div>
  )
}
```
