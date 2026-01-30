import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useSaasAuth } from '@saas-starter/react'

export function ForgotPasswordPage() {
  const { forgotPassword, isLoading, error } = useSaasAuth()

  const [email, setEmail] = useState('')
  const [localError, setLocalError] = useState('')
  const [success, setSuccess] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLocalError('')

    try {
      await forgotPassword(email)
      setSuccess(true)
    } catch (err: unknown) {
      const error = err as { message?: string }
      setLocalError(error.message || 'Failed to send reset email')
    }
  }

  const displayError = localError || error?.message

  if (success) {
    return (
      <div className="auth-container">
        <div className="auth-card">
          <div className="auth-header">
            <h1>Check your email</h1>
            <p>If an account exists for {email}, we sent a password reset link.</p>
          </div>
          <div className="alert alert-success">
            Check your email for the reset link.
          </div>
          <div className="auth-footer">
            <Link to="/login">Back to login</Link>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h1>Forgot password?</h1>
          <p>Enter your email to receive a reset link</p>
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
              placeholder="you@example.com"
              required
            />
          </div>
          <button type="submit" className="btn btn-primary" disabled={isLoading}>
            {isLoading ? 'Sending...' : 'Send reset link'}
          </button>
        </form>

        <div className="auth-footer">
          Remember your password? <Link to="/login">Sign in</Link>
        </div>
      </div>
    </div>
  )
}
