import { useState, useEffect } from 'react'
import { Link, useNavigate, useLocation, useSearchParams } from 'react-router-dom'
import { useSaasAuth } from '@saas-starter/react'

export function VerifyEmailPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams] = useSearchParams()
  const { verifyEmail, isLoading, error } = useSaasAuth()

  const [token, setToken] = useState('')
  const [localError, setLocalError] = useState('')
  const [success, setSuccess] = useState(false)

  const email = (location.state as { email?: string })?.email || ''

  // Check for token in URL
  useEffect(() => {
    const urlToken = searchParams.get('token')
    if (urlToken) {
      setToken(urlToken)
      handleVerify(urlToken)
    }
  }, [searchParams])

  const handleVerify = async (verifyToken?: string) => {
    const tokenToUse = verifyToken || token
    if (!tokenToUse) {
      setLocalError('Please enter verification token')
      return
    }

    setLocalError('')
    try {
      await verifyEmail(tokenToUse)
      setSuccess(true)
      setTimeout(() => navigate('/'), 2000)
    } catch (err: unknown) {
      const error = err as { message?: string }
      setLocalError(error.message || 'Verification failed')
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    handleVerify()
  }

  const displayError = localError || error?.message

  if (success) {
    return (
      <div className="auth-container">
        <div className="auth-card">
          <div className="auth-header">
            <h1>Email Verified!</h1>
            <p>Redirecting to dashboard...</p>
          </div>
          <div className="alert alert-success">
            Your email has been verified successfully.
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h1>Verify your email</h1>
          <p>
            {email
              ? `We sent a verification link to ${email}`
              : 'Enter the verification token from your email'}
          </p>
        </div>

        {displayError && (
          <div className="alert alert-error">{displayError}</div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="token">Verification Token</label>
            <input
              id="token"
              type="text"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="Paste token from email"
              required
            />
          </div>
          <button type="submit" className="btn btn-primary" disabled={isLoading}>
            {isLoading ? 'Verifying...' : 'Verify Email'}
          </button>
        </form>

        <div className="auth-footer">
          <Link to="/login">Back to login</Link>
        </div>
      </div>
    </div>
  )
}
