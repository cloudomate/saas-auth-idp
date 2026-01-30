import { useState, useEffect, FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'

export function RegisterPage() {
  const navigate = useNavigate()
  const { register, isAuthenticated, isLoading, error, clearError } = useAuth()

  const [email, setEmail] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
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
  }, [email, displayName, password, confirmPassword])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setFormError(null)

    // Validation
    if (!email.trim()) {
      setFormError('Email is required')
      return
    }

    // Basic email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(email)) {
      setFormError('Please enter a valid email address')
      return
    }

    if (!password.trim()) {
      setFormError('Password is required')
      return
    }

    if (password.length < 6) {
      setFormError('Password must be at least 6 characters')
      return
    }

    if (password !== confirmPassword) {
      setFormError('Passwords do not match')
      return
    }

    try {
      await register({
        email,
        password,
        displayName: displayName || undefined,
      })
      navigate('/')
    } catch (err) {
      // Error is handled by context
    }
  }

  const displayError = formError || error

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h1>Create Account</h1>
          <p>Sign up for a new account</p>
        </div>

        {displayError && (
          <div className="alert alert-error">{displayError}</div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email *</label>
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
            <label htmlFor="displayName">Display Name (optional)</label>
            <input
              id="displayName"
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="How should we call you?"
              disabled={isLoading}
              autoComplete="name"
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password *</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Create a password"
              disabled={isLoading}
              autoComplete="new-password"
            />
          </div>

          <div className="form-group">
            <label htmlFor="confirmPassword">Confirm Password *</label>
            <input
              id="confirmPassword"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Confirm your password"
              disabled={isLoading}
              autoComplete="new-password"
            />
          </div>

          <button
            type="submit"
            className="btn btn-primary"
            disabled={isLoading}
            style={{ width: '100%', marginTop: '1rem' }}
          >
            {isLoading ? 'Creating account...' : 'Create Account'}
          </button>
        </form>

        <div className="auth-footer" style={{ marginTop: '2rem', textAlign: 'center' }}>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
            Already have an account?{' '}
            <Link to="/login" style={{ color: 'var(--primary-color)' }}>
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
