import { type ReactNode } from 'react'
import { useSaasAuth } from '../hooks/useSaasAuth'

export interface ProtectedRouteProps {
  children: ReactNode
  /** Component to render while loading */
  loadingFallback?: ReactNode
  /** Component to render when not authenticated */
  fallback?: ReactNode
  /** Require user to have a tenant set up */
  requireTenant?: boolean
  /** Component to render when tenant setup is required */
  tenantSetupFallback?: ReactNode
}

/**
 * Route protection component
 *
 * @example
 * ```tsx
 * <ProtectedRoute
 *   fallback={<LoginPage />}
 *   requireTenant={true}
 *   tenantSetupFallback={<OnboardingPage />}
 * >
 *   <DashboardPage />
 * </ProtectedRoute>
 * ```
 */
export function ProtectedRoute({
  children,
  loadingFallback = null,
  fallback = null,
  requireTenant = false,
  tenantSetupFallback = null,
}: ProtectedRouteProps) {
  const { isAuthenticated, isLoading, needsTenantSetup } = useSaasAuth()

  if (isLoading) {
    return <>{loadingFallback}</>
  }

  if (!isAuthenticated) {
    return <>{fallback}</>
  }

  if (requireTenant && needsTenantSetup) {
    return <>{tenantSetupFallback || fallback}</>
  }

  return <>{children}</>
}
