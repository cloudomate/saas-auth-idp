import { createContext, useContext, useCallback, useEffect, useState, useMemo, type ReactNode } from 'react'
import type {
  User,
  AuthError,
  SaasAuthConfig,
  PlanTier,
  Tenant,
  Subscription,
  Plan,
  Workspace,
  AuthResponse,
  OAuthLoginResponse,
  TenantSetupResponse,
  ApiError,
} from '../types'
import { SaasApiClient, createStorage, isTokenExpired, type Storage } from '../utils/api-client'

// ============================================================================
// Context Types
// ============================================================================

interface SaasAuthContextValue {
  // Auth State
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  error: AuthError | null
  needsTenantSetup: boolean

  // Tenant State
  tenant: Tenant | null
  subscription: Subscription | null
  plan: Plan | null
  plans: Plan[]

  // Workspace State
  workspaces: Workspace[]
  currentWorkspace: Workspace | null

  // Auth Actions
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

  // Tenant Actions
  selectPlan: (tier: PlanTier) => Promise<{ tenantCreated: boolean; redirectTo?: string }>
  setupOrganization: (name: string, slug: string, emailDomain?: string) => Promise<Tenant>
  checkSlug: (slug: string) => Promise<boolean>
  refreshTenant: () => Promise<void>
  fetchPlans: () => Promise<void>

  // Workspace Actions
  setCurrentWorkspace: (workspaceId: string) => void
  createWorkspace: (name: string, slug: string) => Promise<Workspace>
  refreshWorkspaces: () => Promise<void>

  // Utilities
  api: SaasApiClient
  storage: Storage
}

const SaasAuthContext = createContext<SaasAuthContextValue | null>(null)

// ============================================================================
// Provider Props
// ============================================================================

interface SaasAuthProviderProps extends SaasAuthConfig {
  children: ReactNode
}

// ============================================================================
// Provider Component
// ============================================================================

export function SaasAuthProvider({
  children,
  apiUrl,
  storagePrefix = 'saas',
  redirectUri = '/auth/callback',
  onAuthStateChange,
}: SaasAuthProviderProps) {
  // Storage and API client
  const storage = useMemo(() => createStorage(storagePrefix), [storagePrefix])
  const api = useMemo(
    () => new SaasApiClient(apiUrl, storagePrefix, () => storage.getItem('token')),
    [apiUrl, storagePrefix, storage]
  )

  // Auth state
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<AuthError | null>(null)
  const [needsTenantSetup, setNeedsTenantSetup] = useState(false)

  // Tenant state
  const [tenant, setTenant] = useState<Tenant | null>(null)
  const [subscription, setSubscription] = useState<Subscription | null>(null)
  const [plans, setPlans] = useState<Plan[]>([])

  // Workspace state
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [currentWorkspaceId, setCurrentWorkspaceId] = useState<string | null>(null)

  // Derived state
  const isAuthenticated = !!user
  const plan = subscription?.plan || null
  const currentWorkspace = workspaces.find((w) => w.id === currentWorkspaceId) || workspaces[0] || null

  // ============================================================================
  // Token Management
  // ============================================================================

  const setToken = useCallback(
    (token: string) => {
      storage.setItem('token', token)
    },
    [storage]
  )

  const clearAuth = useCallback(() => {
    storage.clear(['token', 'user', 'tenant', 'workspace'])
    setUser(null)
    setTenant(null)
    setSubscription(null)
    setWorkspaces([])
    setNeedsTenantSetup(false)
    onAuthStateChange?.(null)
  }, [storage, onAuthStateChange])

  // ============================================================================
  // Auth Actions
  // ============================================================================

  const loginWithProvider = useCallback(
    async (provider: 'google' | 'github', customRedirectUri?: string) => {
      setError(null)
      try {
        const finalRedirectUri = customRedirectUri || `${window.location.origin}${redirectUri}`
        const response = await api.get<OAuthLoginResponse>(
          `/api/v1/auth/social/${provider}/login?redirect_uri=${encodeURIComponent(finalRedirectUri)}&flow=signup`
        )

        // Store state for callback verification
        storage.setItem('oauth_state', response.state)

        // Redirect to OAuth provider
        window.location.href = response.auth_url
      } catch (err) {
        const apiError = err as ApiError
        setError({ code: apiError.error || 'oauth_error', message: apiError.message || 'OAuth login failed' })
        throw err
      }
    },
    [api, storage, redirectUri]
  )

  const loginWithGoogle = useCallback(
    (customRedirectUri?: string) => loginWithProvider('google', customRedirectUri),
    [loginWithProvider]
  )

  const loginWithGithub = useCallback(
    (customRedirectUri?: string) => loginWithProvider('github', customRedirectUri),
    [loginWithProvider]
  )

  const handleOAuthCallback = useCallback(
    async (code: string, state: string) => {
      setIsLoading(true)
      setError(null)
      try {
        // Verify state
        const storedState = storage.getItem('oauth_state')
        if (state !== storedState) {
          throw { error: 'invalid_state', message: 'OAuth state mismatch' }
        }
        storage.removeItem('oauth_state')

        // Exchange code for token
        const response = await api.post<AuthResponse>('/api/v1/auth/social/callback', {
          code,
          state,
        })

        setToken(response.access_token)
        setUser(response.user)
        setNeedsTenantSetup(response.needs_tenant_setup)
        onAuthStateChange?.(response.user)
      } catch (err) {
        const apiError = err as ApiError
        setError({ code: apiError.error || 'callback_error', message: apiError.message || 'OAuth callback failed' })
        throw err
      } finally {
        setIsLoading(false)
      }
    },
    [api, storage, setToken, onAuthStateChange]
  )

  const loginWithEmail = useCallback(
    async (email: string, password: string) => {
      setIsLoading(true)
      setError(null)
      try {
        const response = await api.post<AuthResponse>('/api/v1/auth/login', { email, password })
        setToken(response.access_token)
        setUser(response.user)
        setNeedsTenantSetup(response.needs_tenant_setup)
        onAuthStateChange?.(response.user)
      } catch (err) {
        const apiError = err as ApiError
        setError({ code: apiError.error || 'login_error', message: apiError.message || 'Login failed' })
        throw err
      } finally {
        setIsLoading(false)
      }
    },
    [api, setToken, onAuthStateChange]
  )

  const signup = useCallback(
    async (email: string, password: string, name: string, planTier?: PlanTier) => {
      setIsLoading(true)
      setError(null)
      try {
        await api.post('/api/v1/auth/register', { email, password, name, plan: planTier })
        // User needs to verify email
      } catch (err) {
        const apiError = err as ApiError
        setError({ code: apiError.error || 'signup_error', message: apiError.message || 'Signup failed' })
        throw err
      } finally {
        setIsLoading(false)
      }
    },
    [api]
  )

  const verifyEmail = useCallback(
    async (token: string) => {
      setIsLoading(true)
      setError(null)
      try {
        const response = await api.post<AuthResponse>('/api/v1/auth/verify-email', { token })
        setToken(response.access_token)
        setUser(response.user)
        setNeedsTenantSetup(response.needs_tenant_setup)
        onAuthStateChange?.(response.user)
      } catch (err) {
        const apiError = err as ApiError
        setError({ code: apiError.error || 'verify_error', message: apiError.message || 'Verification failed' })
        throw err
      } finally {
        setIsLoading(false)
      }
    },
    [api, setToken, onAuthStateChange]
  )

  const forgotPassword = useCallback(
    async (email: string) => {
      setError(null)
      try {
        await api.post('/api/v1/auth/forgot-password', { email })
      } catch (err) {
        const apiError = err as ApiError
        setError({ code: apiError.error || 'forgot_error', message: apiError.message || 'Request failed' })
        throw err
      }
    },
    [api]
  )

  const resetPassword = useCallback(
    async (token: string, newPassword: string) => {
      setError(null)
      try {
        await api.post('/api/v1/auth/reset-password', { token, password: newPassword })
      } catch (err) {
        const apiError = err as ApiError
        setError({ code: apiError.error || 'reset_error', message: apiError.message || 'Reset failed' })
        throw err
      }
    },
    [api]
  )

  const refreshUser = useCallback(async () => {
    const token = storage.getItem('token')
    if (!token || isTokenExpired(token)) {
      clearAuth()
      setIsLoading(false)
      return
    }

    try {
      const response = await api.get<User>('/api/v1/auth/me')
      setUser(response)
      setNeedsTenantSetup(!response.tenantId)
      onAuthStateChange?.(response)
    } catch {
      clearAuth()
    } finally {
      setIsLoading(false)
    }
  }, [api, storage, clearAuth, onAuthStateChange])

  const logout = useCallback(() => {
    clearAuth()
  }, [clearAuth])

  // ============================================================================
  // Tenant Actions
  // ============================================================================

  const fetchPlans = useCallback(async () => {
    try {
      const response = await api.get<{ plans: Plan[] }>('/api/v1/tenant/plans')
      setPlans(response.plans)
    } catch (err) {
      console.error('Failed to fetch plans:', err)
    }
  }, [api])

  const selectPlan = useCallback(
    async (tier: PlanTier): Promise<{ tenantCreated: boolean; redirectTo?: string }> => {
      const response = await api.post<{
        tenant_created: boolean
        redirect_to?: string
        access_token?: string
        tenant?: Tenant
      }>('/api/v1/tenant/select-plan', { plan: tier })

      if (response.access_token) {
        setToken(response.access_token)
      }

      if (response.tenant_created && response.tenant) {
        setTenant(response.tenant)
        setNeedsTenantSetup(false)
      }

      return {
        tenantCreated: response.tenant_created,
        redirectTo: response.redirect_to,
      }
    },
    [api, setToken]
  )

  const setupOrganization = useCallback(
    async (name: string, slug: string, emailDomain?: string): Promise<Tenant> => {
      const response = await api.post<TenantSetupResponse>('/api/v1/tenant/setup', {
        org_name: name,
        org_slug: slug,
        email_domain: emailDomain,
      })

      if (response.access_token) {
        setToken(response.access_token)
      }

      setTenant(response.tenant)
      setWorkspaces([response.workspace])
      setNeedsTenantSetup(false)

      return response.tenant
    },
    [api, setToken]
  )

  const checkSlug = useCallback(
    async (slug: string): Promise<boolean> => {
      const response = await api.get<{ available: boolean }>(`/api/v1/tenant/check-slug?slug=${encodeURIComponent(slug)}`)
      return response.available
    },
    [api]
  )

  const refreshTenant = useCallback(async () => {
    if (!user?.tenantId) return

    try {
      const response = await api.get<{ tenant: Tenant; subscription?: Subscription }>('/api/v1/tenant')
      setTenant(response.tenant)
      if (response.subscription) {
        setSubscription(response.subscription)
      }
    } catch (err) {
      console.error('Failed to fetch tenant:', err)
    }
  }, [api, user?.tenantId])

  // ============================================================================
  // Workspace Actions
  // ============================================================================

  const setCurrentWorkspace = useCallback(
    (workspaceId: string) => {
      setCurrentWorkspaceId(workspaceId)
      storage.setItem('workspace', workspaceId)
    },
    [storage]
  )

  const createWorkspace = useCallback(
    async (name: string, slug: string): Promise<Workspace> => {
      const response = await api.post<{ workspace: Workspace }>('/api/v1/workspaces', {
        display_name: name,
        slug,
      })
      setWorkspaces((prev) => [...prev, response.workspace])
      return response.workspace
    },
    [api]
  )

  const refreshWorkspaces = useCallback(async () => {
    if (!tenant) return

    try {
      const response = await api.get<{ workspaces: Workspace[] }>('/api/v1/workspaces')
      setWorkspaces(response.workspaces)

      // Restore current workspace from storage or use first
      const storedWorkspace = storage.getItem('workspace')
      const validWorkspace = response.workspaces.find((w) => w.id === storedWorkspace)
      setCurrentWorkspaceId(validWorkspace?.id || response.workspaces[0]?.id || null)
    } catch (err) {
      console.error('Failed to fetch workspaces:', err)
    }
  }, [api, tenant, storage])

  // ============================================================================
  // Initialization
  // ============================================================================

  useEffect(() => {
    refreshUser()
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (user?.tenantId) {
      refreshTenant()
      refreshWorkspaces()
    }
  }, [user?.tenantId]) // eslint-disable-line react-hooks/exhaustive-deps

  // ============================================================================
  // Context Value
  // ============================================================================

  const value: SaasAuthContextValue = {
    // Auth State
    user,
    isAuthenticated,
    isLoading,
    error,
    needsTenantSetup,

    // Tenant State
    tenant,
    subscription,
    plan,
    plans,

    // Workspace State
    workspaces,
    currentWorkspace,

    // Auth Actions
    loginWithGoogle,
    loginWithGithub,
    handleOAuthCallback,
    loginWithEmail,
    signup,
    verifyEmail,
    forgotPassword,
    resetPassword,
    logout,
    refreshUser,

    // Tenant Actions
    selectPlan,
    setupOrganization,
    checkSlug,
    refreshTenant,
    fetchPlans,

    // Workspace Actions
    setCurrentWorkspace,
    createWorkspace,
    refreshWorkspaces,

    // Utilities
    api,
    storage,
  }

  return <SaasAuthContext.Provider value={value}>{children}</SaasAuthContext.Provider>
}

// ============================================================================
// Hook
// ============================================================================

export function useSaasAuthContext(): SaasAuthContextValue {
  const context = useContext(SaasAuthContext)
  if (!context) {
    throw new Error('useSaasAuthContext must be used within a SaasAuthProvider')
  }
  return context
}
