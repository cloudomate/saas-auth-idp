import { useEffect } from 'react'
import { useSaasAuthContext } from '../contexts/SaasAuthContext'
import type { UseTenantReturn } from '../types'

/**
 * Hook for tenant and subscription management
 *
 * @example
 * ```tsx
 * const {
 *   tenant,
 *   plan,
 *   plans,
 *   selectPlan,
 *   setupOrganization
 * } = useTenant()
 *
 * // Show plan selection
 * {plans.map(plan => (
 *   <button key={plan.id} onClick={() => selectPlan(plan.tier)}>
 *     {plan.name} - ${plan.monthlyPrice}/mo
 *   </button>
 * ))}
 * ```
 */
export function useTenant(): UseTenantReturn {
  const context = useSaasAuthContext()

  // Fetch plans on mount if authenticated
  useEffect(() => {
    if (context.isAuthenticated && context.plans.length === 0) {
      context.fetchPlans()
    }
  }, [context.isAuthenticated]) // eslint-disable-line react-hooks/exhaustive-deps

  return {
    tenant: context.tenant,
    subscription: context.subscription,
    plan: context.plan,
    plans: context.plans,
    isLoading: context.isLoading,
    error: context.error ? { error: context.error.code, message: context.error.message } : null,

    // Actions
    selectPlan: context.selectPlan,
    setupOrganization: context.setupOrganization,
    checkSlug: context.checkSlug,
    refreshTenant: context.refreshTenant,
  }
}
