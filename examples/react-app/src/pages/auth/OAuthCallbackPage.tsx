import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '../../contexts/AuthContext'

export function OAuthCallbackPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { handleCallback } = useAuth()
  const [error, setError] = useState('')

  useEffect(() => {
    const code = searchParams.get('code')
    const state = searchParams.get('state')
    const errorParam = searchParams.get('error')

    if (errorParam) {
      setError(searchParams.get('error_description') || 'Authentication failed')
      return
    }

    if (code) {
      handleCallback(code, state || '')
        .then(() => {
          navigate('/')
        })
        .catch((err: Error) => {
          setError(err.message || 'Authentication failed')
        })
    } else {
      setError('Invalid callback parameters')
    }
  }, [searchParams, handleCallback, navigate])

  if (error) {
    return (
      <div className="auth-container">
        <div className="auth-card">
          <div className="auth-header">
            <h1>Authentication Failed</h1>
          </div>
          <div className="alert alert-error">{error}</div>
          <button
            className="btn btn-primary"
            onClick={() => navigate('/login')}
          >
            Back to Login
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h1>Authenticating...</h1>
          <p>Please wait while we complete your sign in.</p>
        </div>
        <div className="loading">
          <div className="spinner" />
        </div>
      </div>
    </div>
  )
}
