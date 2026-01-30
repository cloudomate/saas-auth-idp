import type { ReactNode } from 'react'

// User types
export interface User {
  id: string
  email: string
  name: string
  picture?: string
  auth_provider: string
  email_verified: boolean
  is_platform_admin?: boolean
  is_root_admin?: boolean
  admin_of_root_id?: string
  selected_plan?: string
  tenantId?: string
  created_at: string
}

// Auth types
export type AuthProvider = 'google' | 'github' | 'local'
export type PlanTier = 'basic' | 'advanced' | 'enterprise'

export interface AuthError {
  code: string
  message: string
}

export interface ApiError {
  error: string
  message: string
  details?: unknown
}

// Hierarchy types
export interface HierarchyLevel {
  name: string
  display_name: string
  plural: string
  url_path: string
  roles: string[]
  is_root: boolean
}

export interface HierarchyConfig {
  root_level: string
  leaf_level: string
  depth: number
  levels: HierarchyLevel[]
}

// Generic container (replaces Tenant/Workspace)
export interface Container {
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
  _level_config?: {
    display_name: string
    plural: string
    roles: string[]
  }
}

// Membership
export interface Membership {
  user_id: string
  email: string
  name: string
  picture?: string
  role: string
  created_at: string
}

// Legacy aliases for backward compatibility
export type Tenant = Container
export type Workspace = Container

// Plan types
export interface Plan {
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

export interface Subscription {
  id: string
  tenant_id: string
  plan_id: string
  status: 'active' | 'cancelled' | 'past_due' | 'trialing'
  current_period_start: string
  current_period_end: string
  plan?: Plan
}

// API Response types
export interface AuthResponse {
  access_token: string
  token_type: string
  expires_in: number
  user: User
  needs_tenant_setup: boolean
}

export interface OAuthLoginResponse {
  auth_url: string
  state: string
}

export interface TenantSetupResponse {
  access_token: string
  tenant: Container
  workspace: Container
}

// SDK Config
export interface SaasAuthConfig {
  apiUrl: string
  storagePrefix?: string
  redirectUri?: string
  onAuthStateChange?: (user: User | null) => void
}

// Auth context types
export interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  error: AuthError | null
  needsTenantSetup: boolean
}

export interface AuthActions {
  loginWithGoogle: (redirectUri?: string) => Promise<void>
  loginWithGithub: (redirectUri?: string) => Promise<void>
  handleOAuthCallback: (code: string, state: string) => Promise<void>
  loginWithEmail: (email: string, password: string) => Promise<void>
  signup: (email: string, password: string, name: string, plan?: PlanTier) => Promise<void>
  verifyEmail: (token: string) => Promise<void>
  forgotPassword: (email: string) => Promise<void>
  resetPassword: (token: string, newPassword: string) => Promise<void>
  logout: () => void
  refreshUser: () => Promise<void>
}

// Hierarchy context types
export interface HierarchyState {
  config: HierarchyConfig | null
  isLoading: boolean
  error: string | null
  // Current selections at each level
  currentContainers: Record<string, Container | null>
}

export interface HierarchyActions {
  // Fetch hierarchy config
  fetchConfig: () => Promise<void>
  // Container operations
  listContainers: (level: string, parentId?: string) => Promise<Container[]>
  createContainer: (level: string, name: string, slug?: string, parentId?: string) => Promise<Container>
  getContainer: (level: string, idOrSlug: string) => Promise<Container>
  deleteContainer: (level: string, id: string) => Promise<void>
  // Member operations
  listMembers: (level: string, containerId: string) => Promise<Membership[]>
  addMember: (level: string, containerId: string, email: string, role?: string) => Promise<void>
  // Selection
  setCurrentContainer: (level: string, container: Container | null) => void
}

// Provider props
export interface SaasAuthProviderProps {
  apiUrl: string
  storagePrefix?: string
  children: ReactNode
  onAuthStateChange?: (user: User | null) => void
  // Hierarchy customization
  hierarchyConfig?: Partial<HierarchyConfig>
}

// Hook return types
export interface UseSaasAuthReturn extends AuthState, AuthActions {}

export interface UseHierarchyReturn extends HierarchyState, HierarchyActions {
  // Convenience getters
  rootLevel: HierarchyLevel | null
  leafLevel: HierarchyLevel | null
  getLevel: (name: string) => HierarchyLevel | undefined
  // Current root container (tenant/organization)
  currentRoot: Container | null
  // Current leaf container (workspace/project)
  currentLeaf: Container | null
}

// Legacy hook return types (for backward compatibility)
export interface UseTenantReturn {
  tenant: Container | null
  subscription: Subscription | null
  plan: Plan | null
  plans: Plan[]
  isLoading: boolean
  error: { error: string; message: string } | null
  selectPlan: (tier: PlanTier) => Promise<{ tenantCreated: boolean; redirectTo?: string }>
  setupOrganization: (name: string, slug: string, emailDomain?: string) => Promise<Container>
  checkSlug: (slug: string) => Promise<boolean>
  refreshTenant: () => Promise<void>
}

export interface UseWorkspacesReturn {
  workspaces: Container[]
  currentWorkspace: Container | null
  isLoading: boolean
  error: { error: string; message: string } | null
  setCurrentWorkspace: (workspaceId: string) => void
  createWorkspace: (name: string, slug: string) => Promise<Container>
  refreshWorkspaces: () => Promise<void>
}

// Generic container hook (recommended)
export interface UseContainersReturn {
  containers: Container[]
  currentContainer: Container | null
  isLoading: boolean
  error: string | null
  setCurrentContainer: (container: Container | null) => void
  create: (name: string, slug?: string, parentId?: string) => Promise<Container>
  delete: (id: string) => Promise<void>
  listMembers: (containerId: string) => Promise<Membership[]>
  addMember: (containerId: string, email: string, role?: string) => Promise<void>
  level: HierarchyLevel
}
