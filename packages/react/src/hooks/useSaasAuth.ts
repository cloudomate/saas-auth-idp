import { useSaasAuthContext } from '../contexts/SaasAuthContext'
import type { UseSaasAuthReturn } from '../types'

/**
 * Hook for authentication operations
 *
 * @example
 * ```tsx
 * const {
 *   user,
 *   isAuthenticated,
 *   loginWithGoogle,
 *   logout
 * } = useSaasAuth()
 *
 * if (!isAuthenticated) {
 *   return <button onClick={() => loginWithGoogle()}>Sign in</button>
 * }
 *
 * return <div>Welcome, {user.name}!</div>
 * ```
 */
export function useSaasAuth(): UseSaasAuthReturn {
  const context = useSaasAuthContext()

  return {
    // State
    user: context.user,
    isAuthenticated: context.isAuthenticated,
    isLoading: context.isLoading,
    error: context.error,
    needsTenantSetup: context.needsTenantSetup,

    // Social Login
    loginWithGoogle: context.loginWithGoogle,
    loginWithGithub: context.loginWithGithub,
    handleOAuthCallback: context.handleOAuthCallback,

    // Email/Password
    loginWithEmail: context.loginWithEmail,
    signup: context.signup,
    verifyEmail: context.verifyEmail,
    forgotPassword: context.forgotPassword,
    resetPassword: context.resetPassword,

    // Session
    logout: context.logout,
    refreshUser: context.refreshUser,
  }
}
