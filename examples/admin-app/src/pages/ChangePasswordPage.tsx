import { useState, FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAdminAuth } from '../contexts/AdminAuthContext'

export function ChangePasswordPage() {
  const navigate = useNavigate()
  const { changePassword, isLoading, error, clearError, logout } = useAdminAuth()

  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [localError, setLocalError] = useState('')
  const [success, setSuccess] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setLocalError('')
    clearError()

    if (!oldPassword) {
      setLocalError('Current password is required')
      return
    }
    if (!newPassword) {
      setLocalError('New password is required')
      return
    }
    if (newPassword.length < 8) {
      setLocalError('New password must be at least 8 characters')
      return
    }
    if (newPassword !== confirmPassword) {
      setLocalError('Passwords do not match')
      return
    }
    if (oldPassword === newPassword) {
      setLocalError('New password must be different from current password')
      return
    }

    try {
      await changePassword(oldPassword, newPassword)
      setSuccess(true)
      // Redirect to dashboard after short delay
      setTimeout(() => navigate('/'), 1500)
    } catch {
      // Error handled by context
    }
  }

  const displayError = localError || error

  if (success) {
    return (
      <div className="login-container">
        <div className="login-card">
          <div className="login-header">
            <h1>Password Changed</h1>
            <p>Your password has been updated successfully.</p>
          </div>
          <div style={{ textAlign: 'center', color: 'var(--success)', marginTop: '1rem' }}>
            Redirecting to dashboard...
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="login-container">
      <div className="login-card">
        <div className="login-header">
          <h1>Change Password</h1>
          <p>You must change your default password before continuing.</p>
        </div>

        {displayError && (
          <div className="alert alert-error">{displayError}</div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="oldPassword">Current Password</label>
            <input
              type="password"
              id="oldPassword"
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              placeholder="Enter current password"
              disabled={isLoading}
              autoComplete="current-password"
              autoFocus
            />
          </div>

          <div className="form-group">
            <label htmlFor="newPassword">New Password</label>
            <input
              type="password"
              id="newPassword"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="Enter new password (min 8 characters)"
              disabled={isLoading}
              autoComplete="new-password"
            />
          </div>

          <div className="form-group">
            <label htmlFor="confirmPassword">Confirm New Password</label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Confirm new password"
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
            {isLoading ? 'Changing Password...' : 'Change Password'}
          </button>
        </form>

        <div style={{ marginTop: '1.5rem', textAlign: 'center' }}>
          <button
            onClick={logout}
            className="btn btn-secondary btn-small"
            style={{ opacity: 0.7 }}
          >
            Sign out
          </button>
        </div>
      </div>
    </div>
  )
}
