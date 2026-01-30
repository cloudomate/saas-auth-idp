import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'

// API URL for auth endpoints
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4455'

// Admin user type
interface AdminUser {
  id: string
  name: string
  displayName: string
  email: string
  avatar: string
  organization: string
  isAdmin: boolean
  isGlobalAdmin: boolean
  passwordChangeRequired: boolean
}

// Admin auth context type
interface AdminAuthContextType {
  user: AdminUser | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  isAdmin: boolean
  requiresPasswordChange: boolean
  error: string | null
  login: (email: string, password: string) => Promise<void>
  changePassword: (oldPassword: string, newPassword: string) => Promise<void>
  logout: () => void
  getToken: () => string | null
  clearError: () => void
}

const AdminAuthContext = createContext<AdminAuthContextType | undefined>(undefined)

// Storage keys - separate from user portal
const TOKEN_KEY = 'admin_auth_token'
const USER_KEY = 'admin_auth_user'

interface AdminAuthProviderProps {
  children: ReactNode
}

export function AdminAuthProvider({ children }: AdminAuthProviderProps) {
  const [user, setUser] = useState<AdminUser | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Load saved session on mount
  useEffect(() => {
    const savedToken = localStorage.getItem(TOKEN_KEY)
    const savedUser = localStorage.getItem(USER_KEY)

    if (savedToken && savedUser) {
      try {
        const parsedUser = JSON.parse(savedUser)
        // Verify user is still an admin
        if (parsedUser.isAdmin) {
          setToken(savedToken)
          setUser(parsedUser)
        } else {
          // Clear invalid admin session
          localStorage.removeItem(TOKEN_KEY)
          localStorage.removeItem(USER_KEY)
        }
      } catch {
        localStorage.removeItem(TOKEN_KEY)
        localStorage.removeItem(USER_KEY)
      }
    }

    setIsLoading(false)
  }, [])

  // Login with email/password - requires admin
  const login = useCallback(async (email: string, password: string) => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await fetch(`${API_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      })

      const data = await response.json()

      if (!response.ok) {
        throw new Error(data.message || 'Login failed')
      }

      const accessToken = data.access_token
      if (!accessToken) {
        throw new Error('No access token in response')
      }

      // Parse JWT to get user info
      const userInfo = parseJwt(accessToken)

      // Check properties for passwordChangeRequired
      const properties = userInfo.properties as Record<string, string> | undefined
      const passwordChangeRequired = properties?.passwordChangeRequired === 'true'

      const adminUser: AdminUser = {
        id: String(userInfo.name || userInfo.sub || email),
        name: String(userInfo.name || email.split('@')[0]),
        displayName: String(userInfo.displayName || userInfo.name || email.split('@')[0]),
        email: String(userInfo.email || email),
        avatar: String(userInfo.avatar || ''),
        organization: String(userInfo.owner || 'saas-platform'),
        isAdmin: Boolean(userInfo.isAdmin),
        isGlobalAdmin: Boolean(userInfo.isGlobalAdmin),
        passwordChangeRequired,
      }

      // CRITICAL: Verify admin privileges
      if (!adminUser.isAdmin) {
        throw new Error('Access denied: Admin privileges required')
      }

      setToken(accessToken)
      setUser(adminUser)
      localStorage.setItem(TOKEN_KEY, accessToken)
      localStorage.setItem(USER_KEY, JSON.stringify(adminUser))
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed'
      setError(message)
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [])

  // Change password - also clears the passwordChangeRequired flag
  const changePassword = useCallback(async (oldPassword: string, newPassword: string) => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await fetch(`${API_URL}/api/v1/auth/change-password`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          old_password: oldPassword,
          new_password: newPassword,
        }),
      })

      const data = await response.json()

      if (!response.ok) {
        throw new Error(data.message || 'Failed to change password')
      }

      // Update user to clear passwordChangeRequired flag
      if (user) {
        const updatedUser = { ...user, passwordChangeRequired: false }
        setUser(updatedUser)
        localStorage.setItem(USER_KEY, JSON.stringify(updatedUser))
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to change password'
      setError(message)
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [token, user])

  // Logout
  const logout = useCallback(() => {
    setToken(null)
    setUser(null)
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }, [])

  // Get current token
  const getToken = useCallback(() => {
    return token
  }, [token])

  // Clear error
  const clearError = useCallback(() => {
    setError(null)
  }, [])

  const value: AdminAuthContextType = {
    user,
    token,
    isAuthenticated: !!token && !!user,
    isLoading,
    isAdmin: user?.isAdmin || false,
    requiresPasswordChange: user?.passwordChangeRequired || false,
    error,
    login,
    changePassword,
    logout,
    getToken,
    clearError,
  }

  return (
    <AdminAuthContext.Provider value={value}>
      {children}
    </AdminAuthContext.Provider>
  )
}

// Hook to use admin auth context
export function useAdminAuth() {
  const context = useContext(AdminAuthContext)
  if (context === undefined) {
    throw new Error('useAdminAuth must be used within an AdminAuthProvider')
  }
  return context
}

// Helper function to parse JWT
function parseJwt(token: string): Record<string, unknown> {
  try {
    const base64Url = token.split('.')[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    )
    return JSON.parse(jsonPayload)
  } catch {
    return {}
  }
}
