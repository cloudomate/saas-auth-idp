import { useState, useEffect, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'

export function LoginPage() {
  const navigate = useNavigate()
  const { login, socialLogin, isAuthenticated, isLoading, error, clearError } = useAuth()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [formError, setFormError] = useState<string | null>(null)

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/')
    }
  }, [isAuthenticated, navigate])

  // Clear errors when inputs change
  useEffect(() => {
    if (formError || error) {
      setFormError(null)
      clearError()
    }
  }, [email, password])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setFormError(null)

    if (!email.trim() || !password.trim()) {
      setFormError('Please enter email and password')
      return
    }

    try {
      await login(email, password)
      navigate('/')
    } catch (err) {
      // Error is handled by context
    }
  }

  const handleSocialLogin = async (provider: string) => {
    try {
      await socialLogin(provider)
    } catch (err) {
      // Error is handled by context
    }
  }

  const displayError = formError || error

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h1>Welcome Back</h1>
          <p>Sign in to your account</p>
        </div>

        {displayError && (
          <div className="alert alert-error">{displayError}</div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Enter your email"
              disabled={isLoading}
              autoComplete="email"
              autoFocus
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              disabled={isLoading}
              autoComplete="current-password"
            />
          </div>

          <button
            type="submit"
            className="btn btn-primary"
            disabled={isLoading}
            style={{ width: '100%', marginTop: '1rem' }}
          >
            {isLoading ? 'Signing in...' : 'Sign In'}
          </button>
        </form>

        <div className="divider" style={{ margin: '1.5rem 0', display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <hr style={{ flex: 1, border: 'none', borderTop: '1px solid var(--border-color)' }} />
          <span style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>or continue with</span>
          <hr style={{ flex: 1, border: 'none', borderTop: '1px solid var(--border-color)' }} />
        </div>

        <div className="social-buttons" style={{ display: 'flex', gap: '0.5rem' }}>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => handleSocialLogin('google')}
            disabled={isLoading}
            style={{ flex: 1 }}
          >
            Google
          </button>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => handleSocialLogin('github')}
            disabled={isLoading}
            style={{ flex: 1 }}
          >
            GitHub
          </button>
        </div>

        <div className="auth-footer" style={{ marginTop: '2rem', textAlign: 'center' }}>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
            Don't have an account?{' '}
            <Link to="/register" style={{ color: 'var(--primary-color)' }}>
              Sign up
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
