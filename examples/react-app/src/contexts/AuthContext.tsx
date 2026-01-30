import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'

// API URL for auth endpoints
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4455'

// User type
interface AuthUser {
  id: string
  name: string
  displayName: string
  email: string
  avatar: string
  organization: string
  isAdmin: boolean
  isGlobalAdmin: boolean
}

// Auth context type
interface AuthContextType {
  user: AuthUser | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null
  login: (email: string, password: string) => Promise<void>
  register: (data: RegisterData) => Promise<void>
  socialLogin: (provider: string) => Promise<void>
  handleCallback: (code: string, state: string) => Promise<void>
  changePassword: (oldPassword: string, newPassword: string) => Promise<void>
  logout: () => void
  getToken: () => string | null
  clearError: () => void
}

interface RegisterData {
  email: string
  password: string
  displayName?: string
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

// Storage keys
const TOKEN_KEY = 'auth_token'
const USER_KEY = 'auth_user'

interface AuthProviderProps {
  children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Load saved session on mount
  useEffect(() => {
    const savedToken = localStorage.getItem(TOKEN_KEY)
    const savedUser = localStorage.getItem(USER_KEY)

    if (savedToken && savedUser) {
      try {
        setToken(savedToken)
        setUser(JSON.parse(savedUser))
      } catch {
        localStorage.removeItem(TOKEN_KEY)
        localStorage.removeItem(USER_KEY)
      }
    }

    setIsLoading(false)
  }, [])

  // Login with email/password
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

      const authUser: AuthUser = {
        id: String(userInfo.name || userInfo.sub || email),
        name: String(userInfo.name || email.split('@')[0]),
        displayName: String(userInfo.displayName || userInfo.name || email.split('@')[0]),
        email: String(userInfo.email || email),
        avatar: String(userInfo.avatar || ''),
        organization: String(userInfo.owner || 'built-in'),
        isAdmin: Boolean(userInfo.isAdmin),
        isGlobalAdmin: Boolean(userInfo.isGlobalAdmin),
      }

      setToken(accessToken)
      setUser(authUser)
      localStorage.setItem(TOKEN_KEY, accessToken)
      localStorage.setItem(USER_KEY, JSON.stringify(authUser))
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed'
      setError(message)
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [])

  // Register new user
  const register = useCallback(async (data: RegisterData) => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await fetch(`${API_URL}/api/v1/auth/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: data.email,
          password: data.password,
          display_name: data.displayName,
        }),
      })

      const result = await response.json()

      if (!response.ok) {
        throw new Error(result.message || 'Registration failed')
      }

      // Auto-login after registration
      await login(data.email, data.password)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Registration failed'
      setError(message)
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [login])

  // Social login - get URL and redirect
  const socialLogin = useCallback(async (provider: string) => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await fetch(`${API_URL}/api/v1/auth/social/${provider}`)
      const data = await response.json()

      if (!response.ok || !data.url) {
        throw new Error(data.message || 'Failed to get login URL')
      }

      // Redirect to social provider
      window.location.href = data.url
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Social login failed'
      setError(message)
      setIsLoading(false)
      throw err
    }
  }, [])

  // Handle OAuth callback
  const handleCallback = useCallback(async (code: string, _state: string) => {
    setIsLoading(true)
    setError(null)

    try {
      const response = await fetch(`${API_URL}/api/v1/auth/callback`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ code }),
      })

      const data = await response.json()

      if (!response.ok) {
        throw new Error(data.message || 'Callback failed')
      }

      const accessToken = data.access_token
      if (!accessToken) {
        throw new Error('No access token in response')
      }

      // Parse JWT to get user info
      const userInfo = parseJwt(accessToken)

      const authUser: AuthUser = {
        id: String(userInfo.name || userInfo.sub || ''),
        name: String(userInfo.name || ''),
        displayName: String(userInfo.displayName || userInfo.name || ''),
        email: String(userInfo.email || ''),
        avatar: String(userInfo.avatar || ''),
        organization: String(userInfo.owner || 'built-in'),
        isAdmin: Boolean(userInfo.isAdmin),
        isGlobalAdmin: Boolean(userInfo.isGlobalAdmin),
      }

      setToken(accessToken)
      setUser(authUser)
      localStorage.setItem(TOKEN_KEY, accessToken)
      localStorage.setItem(USER_KEY, JSON.stringify(authUser))
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Authentication failed'
      setError(message)
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [])

  // Change password
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
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to change password'
      setError(message)
      throw err
    } finally {
      setIsLoading(false)
    }
  }, [token])

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

  const value: AuthContextType = {
    user,
    token,
    isAuthenticated: !!token && !!user,
    isLoading,
    error,
    login,
    register,
    socialLogin,
    handleCallback,
    changePassword,
    logout,
    getToken,
    clearError,
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}

// Hook to use auth context
export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
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
