import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTenant, useSaasAuth } from '@saas-starter/react'
import type { Plan } from '@saas-starter/react'

export function SelectPlanPage() {
  const navigate = useNavigate()
  const { user } = useSaasAuth()
  const { plans, selectPlan, isLoading, error } = useTenant()

  const [selectedPlan, setSelectedPlan] = useState<string>('')
  const [localError, setLocalError] = useState('')

  // If user already has a tenant, redirect to dashboard
  useEffect(() => {
    if (user?.tenantId) {
      navigate('/')
    }
  }, [user, navigate])

  const handleSelectPlan = async (plan: Plan) => {
    setSelectedPlan(plan.tier)
    setLocalError('')

    try {
      await selectPlan(plan.tier)

      // Basic plan auto-creates tenant, others need org setup
      if (plan.tier === 'basic') {
        navigate('/')
      } else {
        navigate('/onboarding/setup')
      }
    } catch (err: unknown) {
      const error = err as { message?: string }
      setLocalError(error.message || 'Failed to select plan')
    }
  }

  const displayError = localError || (error ? error.message : null)

  return (
    <div className="auth-container" style={{ alignItems: 'flex-start', paddingTop: '4rem' }}>
      <div style={{ width: '100%', maxWidth: '900px' }}>
        <div className="auth-header">
          <h1>Choose your plan</h1>
          <p>Select the plan that best fits your needs</p>
        </div>

        {displayError && (
          <div className="alert alert-error">{displayError}</div>
        )}

        <div className="plans-grid">
          {plans.map((plan) => (
            <div
              key={plan.id}
              className={`plan-card ${selectedPlan === plan.tier ? 'selected' : ''}`}
              onClick={() => !isLoading && handleSelectPlan(plan)}
              style={{ cursor: isLoading ? 'not-allowed' : 'pointer' }}
            >
              <div className="plan-name">{plan.name}</div>
              <div className="plan-price">
                ${plan.monthly_price}
                <span>/month</span>
              </div>
              <p style={{ color: 'var(--text-secondary)', margin: '0.5rem 0' }}>
                {plan.description}
              </p>
              <ul className="plan-features">
                {plan.features.map((feature, index) => (
                  <li key={index}>{feature}</li>
                ))}
              </ul>
              <button
                className={`btn ${selectedPlan === plan.tier ? 'btn-primary' : 'btn-secondary'}`}
                disabled={isLoading}
              >
                {isLoading && selectedPlan === plan.tier
                  ? 'Selecting...'
                  : `Select ${plan.name}`}
              </button>
            </div>
          ))}
        </div>

        {plans.length === 0 && !isLoading && (
          <div className="loading">
            <p>Loading plans...</p>
          </div>
        )}
      </div>
    </div>
  )
}
