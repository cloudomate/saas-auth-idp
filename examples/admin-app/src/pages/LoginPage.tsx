import { useState, useEffect, FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAdminAuth } from '../contexts/AdminAuthContext'

export function LoginPage() {
  const navigate = useNavigate()
  const { login, isAuthenticated, isLoading, error, clearError } = useAdminAuth()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [localError, setLocalError] = useState('')

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate('/')
    }
  }, [isAuthenticated, navigate])

  // Clear errors when inputs change
  useEffect(() => {
    if (error) clearError()
    if (localError) setLocalError('')
  }, [email, password])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setLocalError('')

    if (!email.trim()) {
      setLocalError('Email is required')
      return
    }
    if (!password) {
      setLocalError('Password is required')
      return
    }

    try {
      await login(email, password)
      navigate('/')
    } catch {
      // Error is handled by context
    }
  }

  const displayError = localError || error

  return (
    <div className="login-container">
      <div className="login-card">
        <div className="login-header">
          <h1>Admin Portal</h1>
          <p>Sign in with your platform admin account</p>
        </div>

        {displayError && (
          <div className="alert alert-error">{displayError}</div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="admin@localhost"
              disabled={isLoading}
              autoComplete="email"
              autoFocus
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
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

        <p style={{ textAlign: 'center', color: 'var(--text-secondary)', fontSize: '0.875rem', marginTop: '1.5rem' }}>
          Only users with <strong>Admin</strong> privileges can access this portal.
        </p>

        <div style={{ marginTop: '1.5rem', padding: '1rem', background: 'var(--bg-tertiary)', borderRadius: '8px', fontSize: '0.875rem' }}>
          <p style={{ margin: 0, color: 'var(--text-secondary)' }}>
            <strong>Default Admin:</strong><br />
            Email: admin@localhost<br />
            Password: admin123
          </p>
          <p style={{ margin: '0.5rem 0 0 0', color: 'var(--warning)', fontSize: '0.8rem' }}>
            Change the default password after first login!
          </p>
        </div>
      </div>
    </div>
  )
}
