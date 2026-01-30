# Authentication Guide

This guide covers all authentication methods and flows supported by the system.

## Overview

The system supports three authentication methods:
1. **OAuth 2.0** (Google, GitHub)
2. **Email/Password** with email verification
3. **API Keys** for programmatic access

## OAuth Authentication

### Supported Providers

| Provider | Scopes | User Data Retrieved |
|----------|--------|---------------------|
| Google | `email`, `profile` | Email, name, profile picture |
| GitHub | `user:email` | Email, name/login, avatar |

### OAuth Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │     │  Backend │     │ Provider │     │ Database │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │ 1. loginWithGoogle()           │                │
     │───────────────▶│                │                │
     │                │                │                │
     │                │ 2. Generate state               │
     │                │───────────────────────────────▶│
     │                │                │                │
     │ 3. auth_url + state            │                │
     │◀───────────────│                │                │
     │                │                │                │
     │ 4. Redirect to auth_url        │                │
     │────────────────────────────────▶│                │
     │                │                │                │
     │ 5. User authorizes             │                │
     │◀────────────────────────────────│                │
     │                │                │                │
     │ 6. Redirect with code + state  │                │
     │────────────────────────────────▶│                │
     │                │                │                │
     │ 7. POST /callback {code, state}│                │
     │───────────────▶│                │                │
     │                │                │                │
     │                │ 8. Verify state│                │
     │                │◀───────────────────────────────│
     │                │                │                │
     │                │ 9. Exchange code for token     │
     │                │───────────────▶│                │
     │                │                │                │
     │                │ 10. Get user info              │
     │                │◀───────────────│                │
     │                │                │                │
     │                │ 11. Create/update user         │
     │                │───────────────────────────────▶│
     │                │                │                │
     │ 12. JWT token + user           │                │
     │◀───────────────│                │                │
```

### Implementation (React SDK)

```tsx
import { useSaasAuth } from '@saas-starter/react'

function LoginPage() {
  const { loginWithGoogle, loginWithGithub, handleOAuthCallback } = useSaasAuth()

  // Step 1: Initiate OAuth
  const handleGoogleLogin = async () => {
    await loginWithGoogle()
    // Browser redirects to Google
  }

  // Step 2: Handle callback (on callback page)
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const code = params.get('code')
    const state = params.get('state')

    if (code && state) {
      handleOAuthCallback(code, state)
    }
  }, [])

  return (
    <button onClick={handleGoogleLogin}>
      Sign in with Google
    </button>
  )
}
```

### Pre-selecting a Plan

Pass a plan tier during OAuth to auto-select it:

```tsx
// The SDK handles this automatically when using the provider
await loginWithGoogle('/callback?plan=advanced')

// Or manually via API
GET /api/v1/auth/social/google/login?plan=advanced
```

### Backend Configuration

Set OAuth credentials in environment:

```bash
# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret

# GitHub OAuth
GITHUB_CLIENT_ID=your-client-id
GITHUB_CLIENT_SECRET=your-client-secret
```

**Redirect URI Configuration**:
- Google: `https://console.cloud.google.com/apis/credentials`
- GitHub: `https://github.com/settings/developers`

Set redirect URI to: `http://localhost:4455/api/v1/auth/social/callback`

## Email/Password Authentication

### Registration Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │     │  Backend │     │ Database │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     │ 1. signup(email, password, name)│
     │───────────────▶│                │
     │                │                │
     │                │ 2. Hash password
     │                │ 3. Generate verify token
     │                │                │
     │                │ 4. Create user │
     │                │───────────────▶│
     │                │                │
     │ 5. Success + verify_token      │
     │◀───────────────│                │
     │                │                │
     │ (Email with verify link sent)  │
     │                │                │
     │ 6. verifyEmail(token)          │
     │───────────────▶│                │
     │                │                │
     │                │ 7. Validate token
     │                │ 8. Mark verified
     │                │───────────────▶│
     │                │                │
     │ 9. JWT token + user            │
     │◀───────────────│                │
```

### Implementation

```tsx
import { useSaasAuth } from '@saas-starter/react'

function RegistrationPage() {
  const { signup, verifyEmail, error } = useSaasAuth()
  const [verifyToken, setVerifyToken] = useState('')

  const handleSignup = async (e) => {
    e.preventDefault()
    try {
      const result = await signup(email, password, name)
      // In development, token is returned directly
      // In production, sent via email
      setVerifyToken(result.verify_token)
    } catch (err) {
      console.error('Signup failed:', err.message)
    }
  }

  const handleVerify = async () => {
    try {
      await verifyEmail(verifyToken)
      // User is now logged in
    } catch (err) {
      console.error('Verification failed:', err.message)
    }
  }

  return (
    <form onSubmit={handleSignup}>
      {error && <div className="error">{error}</div>}
      <input type="email" value={email} onChange={...} />
      <input type="password" value={password} onChange={...} />
      <input type="text" value={name} onChange={...} />
      <button type="submit">Sign Up</button>
    </form>
  )
}
```

### Login Flow

```tsx
function LoginPage() {
  const { loginWithEmail, error } = useSaasAuth()

  const handleLogin = async (e) => {
    e.preventDefault()
    try {
      await loginWithEmail(email, password)
      // User is now logged in
    } catch (err) {
      if (err.error === 'email_not_verified') {
        // Redirect to verification page
      }
    }
  }

  return (
    <form onSubmit={handleLogin}>
      {error && <div className="error">{error}</div>}
      <input type="email" value={email} onChange={...} />
      <input type="password" value={password} onChange={...} />
      <button type="submit">Login</button>
    </form>
  )
}
```

### Password Reset Flow

```tsx
function PasswordResetPage() {
  const { forgotPassword, resetPassword } = useSaasAuth()
  const [step, setStep] = useState<'request' | 'reset'>('request')

  const handleForgotPassword = async () => {
    const result = await forgotPassword(email)
    // In dev, token returned; in prod, sent via email
    setStep('reset')
  }

  const handleReset = async () => {
    await resetPassword(token, newPassword)
    // Password reset, redirect to login
  }

  if (step === 'request') {
    return (
      <form onSubmit={handleForgotPassword}>
        <input type="email" value={email} onChange={...} />
        <button>Send Reset Link</button>
      </form>
    )
  }

  return (
    <form onSubmit={handleReset}>
      <input type="password" value={newPassword} onChange={...} />
      <button>Reset Password</button>
    </form>
  )
}
```

## API Key Authentication

API keys provide programmatic access to the API, scoped to a specific workspace.

### API Key Format

```
sk-<key_id>-<secret>
```

Example: `sk-abc123-def456ghi789...`

### Using API Keys

```bash
curl -H "Authorization: Bearer sk-abc123-def456ghi789..." \
     -H "X-Workspace-ID: 990e8400-e29b-41d4-a716-446655440001" \
     http://localhost:4455/api/v1/resources
```

### Key Permissions

API keys inherit permissions from the workspace:
- `can_read`: Read-only access
- `can_write`: Read and write access

API keys cannot perform management operations (delete workspace, manage members).

## JWT Token Structure

### Token Claims

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "type": "platform",
  "email_verified": true,
  "is_tenant_admin": true,
  "tenant_id": "660e8400-e29b-41d4-a716-446655440001",
  "iat": 1705312200,
  "exp": 1705398600
}
```

### Token Lifetime

- **Access Token**: 24 hours
- **Verification Token**: 24 hours
- **Password Reset Token**: 1 hour
- **OAuth State**: 10 minutes

### Token Storage

The SDK stores tokens in localStorage:

```javascript
// Stored as:
localStorage.setItem('myapp_token', 'eyJhbGciOiJIUzI1NiIs...')
```

## Session Management

### Checking Authentication

```tsx
function App() {
  const { isAuthenticated, isLoading, user } = useSaasAuth()

  if (isLoading) {
    return <LoadingSpinner />
  }

  if (!isAuthenticated) {
    return <LoginPage />
  }

  return <Dashboard user={user} />
}
```

### Refreshing User Data

```tsx
const { refreshUser } = useSaasAuth()

// After profile update
await updateProfile(data)
await refreshUser()
```

### Logout

```tsx
const { logout } = useSaasAuth()

const handleLogout = () => {
  logout()
  // Token removed, user state cleared
  navigate('/login')
}
```

## Security Best Practices

### 1. Use HTTPS in Production

Always use HTTPS to prevent token interception.

### 2. Validate Email Before Access

```tsx
function ProtectedContent() {
  const { user } = useSaasAuth()

  if (!user.email_verified) {
    return <VerifyEmailPrompt />
  }

  return <Content />
}
```

### 3. Handle Token Expiration

```tsx
// The SDK handles this automatically
// On 401 response, user is logged out

// For manual handling:
import { isTokenExpired } from '@saas-starter/react'

if (isTokenExpired(token)) {
  logout()
}
```

### 4. Secure Password Requirements

Backend enforces:
- Minimum 8 characters
- Hashed with bcrypt (cost factor 10)

### 5. Rate Limiting

Implement rate limiting on auth endpoints in production:
- Login: 5 attempts per minute
- Registration: 3 per minute
- Password reset: 3 per hour

## Development Mode

For local development without OAuth setup:

```bash
# .env
DEV_MODE=true
```

In dev mode:
- AuthZ service bypasses token validation
- Injects default user headers:
  - `X-User-ID: dev-user-123`
  - `X-Tenant-ID: dev-tenant-456`
  - `X-Is-Platform-Admin: true`

## Troubleshooting

### "Provider not configured" Error

OAuth credentials not set in environment:
```bash
GOOGLE_CLIENT_ID=...
GOOGLE_CLIENT_SECRET=...
```

### "Invalid state" Error

OAuth state expired or already used. User should restart login flow.

### "Email not verified" Error

User needs to verify email before logging in:
```tsx
if (error.error === 'email_not_verified') {
  redirect('/verify-email')
}
```

### Token Expired

JWT token has expired. User needs to log in again:
```tsx
// SDK handles this automatically
// Manual handling:
if (error.status === 401) {
  logout()
  redirect('/login')
}
```
