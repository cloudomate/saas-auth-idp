import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTenant, useSaasAuth } from '@saas-starter/react'

export function SetupOrgPage() {
  const navigate = useNavigate()
  const { user } = useSaasAuth()
  const { setupOrganization, checkSlug, isLoading, error } = useTenant()

  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [slugAvailable, setSlugAvailable] = useState(true)
  const [slugChecking, setSlugChecking] = useState(false)
  const [localError, setLocalError] = useState('')

  // If user already has a tenant, redirect to dashboard
  useEffect(() => {
    if (user?.tenantId) {
      navigate('/')
    }
  }, [user, navigate])

  // Auto-generate slug from name
  useEffect(() => {
    if (name && !slug) {
      const generatedSlug = name
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-|-$/g, '')
      setSlug(generatedSlug)
    }
  }, [name])

  // Check slug availability with debounce
  useEffect(() => {
    if (slug.length < 3) {
      setSlugAvailable(true)
      return
    }

    const timer = setTimeout(async () => {
      setSlugChecking(true)
      try {
        const available = await checkSlug(slug)
        setSlugAvailable(available)
      } catch {
        // Ignore errors in slug check
      } finally {
        setSlugChecking(false)
      }
    }, 500)

    return () => clearTimeout(timer)
  }, [slug, checkSlug])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLocalError('')

    if (!name.trim()) {
      setLocalError('Organization name is required')
      return
    }

    if (slug.length < 3) {
      setLocalError('Slug must be at least 3 characters')
      return
    }

    if (!slugAvailable) {
      setLocalError('This slug is not available')
      return
    }

    try {
      await setupOrganization(name, slug)
      navigate('/')
    } catch (err: unknown) {
      const error = err as { message?: string }
      setLocalError(error.message || 'Failed to create organization')
    }
  }

  const displayError = localError || (error ? error.message : null)

  return (
    <div className="auth-container">
      <div className="auth-card" style={{ maxWidth: '500px' }}>
        <div className="auth-header">
          <h1>Set up your organization</h1>
          <p>Create your organization to get started</p>
        </div>

        {displayError && (
          <div className="alert alert-error">{displayError}</div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="name">Organization Name</label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Acme Inc"
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="slug">URL Slug</label>
            <input
              id="slug"
              type="text"
              value={slug}
              onChange={(e) => setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
              placeholder="acme-inc"
              className={!slugAvailable ? 'error' : ''}
              required
            />
            <div style={{ marginTop: '0.25rem', fontSize: '0.875rem' }}>
              {slugChecking ? (
                <span style={{ color: 'var(--text-secondary)' }}>Checking...</span>
              ) : slug.length >= 3 ? (
                slugAvailable ? (
                  <span style={{ color: 'var(--success)' }}>✓ Available</span>
                ) : (
                  <span style={{ color: 'var(--danger)' }}>✗ Not available</span>
                )
              ) : null}
            </div>
            <small style={{ color: 'var(--text-secondary)' }}>
              Your workspace URL will be: app.example.com/{slug || 'your-slug'}
            </small>
          </div>

          <button
            type="submit"
            className="btn btn-primary"
            disabled={isLoading || !slugAvailable || slugChecking}
          >
            {isLoading ? 'Creating...' : 'Create Organization'}
          </button>
        </form>

        <div className="auth-footer">
          <button
            type="button"
            onClick={() => navigate('/onboarding/plan')}
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--primary)',
              cursor: 'pointer',
            }}
          >
            ← Back to plans
          </button>
        </div>
      </div>
    </div>
  )
}
