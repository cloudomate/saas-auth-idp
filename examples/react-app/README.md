# SaaS Auth Example App

A complete example React application demonstrating how to integrate with the SaaS Auth IDP system using the `@saas-starter/react` SDK.

## Features Demonstrated

- **Authentication**
  - OAuth login (Google, GitHub)
  - Email/password registration and login
  - Email verification flow
  - Password reset flow
  - OAuth callback handling

- **Onboarding**
  - Plan selection page
  - Organization setup

- **Protected Routes**
  - Route guards with `ProtectedRoute` component
  - Tenant requirement checks
  - Automatic redirects

- **Dashboard**
  - User and organization info display
  - Workspace switching

- **Workspace Management**
  - List workspaces
  - Create new workspaces
  - View workspace details
  - Manage workspace members

- **ReBAC Demo (Documents Page)**
  - Create documents with different visibility levels
  - Share documents with specific users
  - Permission-based UI (edit/share/delete buttons)
  - Role-based access (owner, editor, viewer)

- **ABAC Demo (Projects Page)**
  - Environment-based policies (production, staging, dev)
  - Status-based rules (active, paused, archived)
  - Admin toggle to see permission changes
  - Policy violation explanations

- **Settings**
  - Profile information
  - Organization details
  - Subscription information

## Prerequisites

- Node.js 18+
- Backend services running (see main README)
- Sample API running (for ReBAC/ABAC demos) - see `examples/sample-api/`

## Getting Started

### 1. Start the Backend

Make sure the backend services are running:

```bash
# From the project root
docker compose up -d
```

### 2. Install Dependencies

```bash
cd examples/react-app
npm install
```

### 3. Configure Environment

Create a `.env` file:

```bash
cp .env.example .env
```

Edit `.env` if your API is running on a different URL:

```env
VITE_API_URL=http://localhost:4455
```

### 4. Start the Development Server

```bash
npm run dev
```

The app will be available at `http://localhost:3000`.

## Project Structure

```
src/
├── main.tsx                    # Entry point with SaasAuthProvider
├── App.tsx                     # Routes configuration
├── styles/
│   └── index.css              # Global styles
├── components/
│   └── layout/
│       └── AppLayout.tsx      # Main app layout with sidebar
└── pages/
    ├── auth/
    │   ├── LoginPage.tsx      # Login with OAuth & email
    │   ├── RegisterPage.tsx   # Email registration
    │   ├── VerifyEmailPage.tsx
    │   ├── ForgotPasswordPage.tsx
    │   ├── ResetPasswordPage.tsx
    │   └── OAuthCallbackPage.tsx
    ├── onboarding/
    │   ├── SelectPlanPage.tsx # Plan selection
    │   └── SetupOrgPage.tsx   # Organization setup
    └── app/
        ├── DashboardPage.tsx  # Main dashboard
        ├── WorkspacesPage.tsx # Workspace list
        ├── WorkspaceDetailPage.tsx # Workspace details & members
        └── SettingsPage.tsx   # User & org settings
```

## Key Integration Points

### 1. Provider Setup

```tsx
// main.tsx
import { SaasAuthProvider } from '@saas-starter/react'

<SaasAuthProvider
  apiUrl="http://localhost:4455"
  storagePrefix="example-app"
>
  <App />
</SaasAuthProvider>
```

### 2. Authentication Hooks

```tsx
// Using useSaasAuth for authentication
const {
  user,
  isAuthenticated,
  isLoading,
  loginWithGoogle,
  loginWithEmail,
  logout,
} = useSaasAuth()
```

### 3. Tenant Management

```tsx
// Using useTenant for organization management
const {
  tenant,
  plans,
  selectPlan,
  setupOrganization,
} = useTenant()
```

### 4. Workspace Management

```tsx
// Using useWorkspaces for workspace operations
const {
  workspaces,
  currentWorkspace,
  setCurrentWorkspace,
  createWorkspace,
} = useWorkspaces()
```

### 5. Protected Routes

```tsx
<ProtectedRoute
  fallback={<Navigate to="/login" />}
  requireTenant
  tenantSetupComponent={<Navigate to="/onboarding/plan" />}
>
  <DashboardPage />
</ProtectedRoute>
```

## Development Mode

If you're running the backend in development mode (`DEV_MODE=true`), authentication is bypassed and you'll be automatically logged in with a dev user.

## Build for Production

```bash
npm run build
```

The built files will be in the `dist/` directory.

## Customization

### Styling

All styles are in `src/styles/index.css`. The app uses CSS variables for theming:

```css
:root {
  --primary: #6366f1;
  --primary-dark: #4f46e5;
  --secondary: #64748b;
  --success: #22c55e;
  --danger: #ef4444;
  /* ... */
}
```

### Adding New Pages

1. Create the page component in `src/pages/`
2. Add the route in `src/App.tsx`
3. Use the SDK hooks to interact with the backend

### Using Custom Hierarchy

If you've configured a custom hierarchy (e.g., tenant → team → project), you can use the `useHierarchy` hook:

```tsx
const {
  config,
  listContainers,
  createContainer,
  setCurrentContainer,
} = useHierarchy()

// List teams
const teams = await listContainers('team', tenantId)

// Create a project
const project = await createContainer('project', 'My Project', 'my-project', teamId)
```

## Troubleshooting

### CORS Errors

Make sure the frontend URL is allowed in the backend CORS configuration:
- Default allowed: `http://localhost:3000`

### OAuth Not Working

1. Check that OAuth providers are configured in the backend `.env`
2. Verify redirect URIs match in provider settings

### "Unauthorized" Errors

1. Check that the backend is running
2. Verify the API URL in `.env`
3. Try clearing localStorage and logging in again

## Running with Sample API (ReBAC/ABAC Demo)

To see the Documents (ReBAC) and Projects (ABAC) demos in action:

```bash
# Terminal 1: Start auth services
docker compose up -d

# Terminal 2: Start sample API
cd examples/sample-api
go run main.go

# Terminal 3: Start React app
cd examples/react-app
npm install
npm run dev
```

Then navigate to:
- **Documents**: http://localhost:3000/documents (ReBAC demo)
- **Projects**: http://localhost:3000/projects (ABAC demo)

## Learn More

- [React SDK Reference](../../docs/react-sdk.md)
- [API Reference](../../docs/api-reference.md)
- [Authentication Guide](../../docs/authentication.md)
- [ReBAC & ABAC Guide](../../docs/rebac-abac-guide.md)
