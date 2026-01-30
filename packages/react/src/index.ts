// ============================================================================
// @saas-starter/react
// Multi-tenant SaaS authentication SDK for React
// ============================================================================

// Context & Provider
export { SaasAuthProvider, useSaasAuthContext } from './contexts/SaasAuthContext'

// Hooks
export { useSaasAuth } from './hooks/useSaasAuth'
export { useTenant } from './hooks/useTenant'
export { useWorkspaces } from './hooks/useWorkspaces'
export { useHierarchy } from './hooks/useHierarchy'
export { useContainers } from './hooks/useContainers'

// Components
export { ProtectedRoute, type ProtectedRouteProps } from './components/ProtectedRoute'
export { SocialLoginButtons, type SocialLoginButtonsProps } from './components/SocialLoginButtons'

// Types
export type {
  // User & Auth
  User,
  AuthState,
  AuthError,
  AuthProvider,
  ApiError,

  // Hierarchy (pluggable multi-level support)
  HierarchyConfig,
  HierarchyLevel,
  Container,

  // Legacy aliases (backward compatible)
  Tenant,
  Workspace,

  // Plans & Subscriptions
  PlanTier,
  Plan,
  Subscription,

  // Membership
  Membership,

  // API Responses
  AuthResponse,
  OAuthLoginResponse,
  TenantSetupResponse,

  // Config
  SaasAuthConfig,
  SaasAuthProviderProps,

  // Hook Returns
  UseSaasAuthReturn,
  UseTenantReturn,
  UseWorkspacesReturn,
  UseHierarchyReturn,
  UseContainersReturn,
} from './types'

// Utilities (for advanced use cases)
export { SaasApiClient, createStorage, decodeJwt, isTokenExpired } from './utils/api-client'
