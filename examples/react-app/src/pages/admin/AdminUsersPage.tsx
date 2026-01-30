import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'

interface User {
  id: string
  email: string
  name: string
  picture?: string
  is_platform_admin: boolean
  created_at: string
  updated_at: string
}

export function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8001'

  useEffect(() => {
    fetchUsers()
  }, [])

  const fetchUsers = async () => {
    try {
      const response = await fetch(`${apiUrl}/api/v1/admin/users`, {
        headers: {
          'X-User-ID': 'user-1',
          'X-Is-Platform-Admin': 'true',
        },
      })

      if (!response.ok) {
        throw new Error('Failed to fetch users')
      }

      const data = await response.json()
      setUsers(data.users || [])
    } catch (err) {
      setError('Failed to load users')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const toggleAdmin = async (userId: string, currentStatus: boolean) => {
    setActionLoading(userId)
    try {
      const response = await fetch(`${apiUrl}/api/v1/admin/users/${userId}/admin`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': 'user-1',
          'X-Is-Platform-Admin': 'true',
        },
        body: JSON.stringify({ is_platform_admin: !currentStatus }),
      })

      if (!response.ok) {
        throw new Error('Failed to update user')
      }

      // Refresh users list
      await fetchUsers()
    } catch (err) {
      setError('Failed to update user admin status')
      console.error(err)
    } finally {
      setActionLoading(null)
    }
  }

  const deleteUser = async (userId: string) => {
    if (!confirm('Are you sure you want to delete this user?')) {
      return
    }

    setActionLoading(userId)
    try {
      const response = await fetch(`${apiUrl}/api/v1/admin/users/${userId}`, {
        method: 'DELETE',
        headers: {
          'X-User-ID': 'user-1',
          'X-Is-Platform-Admin': 'true',
        },
      })

      if (!response.ok) {
        const data = await response.json()
        throw new Error(data.error || 'Failed to delete user')
      }

      // Refresh users list
      await fetchUsers()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete user')
      console.error(err)
    } finally {
      setActionLoading(null)
    }
  }

  if (loading) {
    return (
      <div className="admin-container">
        <div className="loading">Loading users...</div>
      </div>
    )
  }

  return (
    <div className="admin-container">
      <div className="admin-header">
        <div className="header-with-back">
          <Link to="/admin" className="back-link">&larr; Back to Dashboard</Link>
          <h1>User Management</h1>
        </div>
        <p>Manage platform users and their admin privileges</p>
      </div>

      {error && (
        <div className="alert alert-error" style={{ marginBottom: '1rem' }}>
          {error}
          <button onClick={() => setError('')} className="alert-close">&times;</button>
        </div>
      )}

      <div className="admin-table-container">
        <table className="admin-table">
          <thead>
            <tr>
              <th>User</th>
              <th>Email</th>
              <th>Role</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user) => (
              <tr key={user.id}>
                <td>
                  <div className="user-cell">
                    {user.picture ? (
                      <img src={user.picture} alt={user.name} className="user-avatar" />
                    ) : (
                      <div className="user-avatar-placeholder">
                        {user.name.charAt(0).toUpperCase()}
                      </div>
                    )}
                    <span>{user.name}</span>
                  </div>
                </td>
                <td>{user.email}</td>
                <td>
                  <span className={`badge ${user.is_platform_admin ? 'badge-admin' : 'badge-user'}`}>
                    {user.is_platform_admin ? 'Platform Admin' : 'User'}
                  </span>
                </td>
                <td>{new Date(user.created_at).toLocaleDateString()}</td>
                <td>
                  <div className="action-buttons-small">
                    <button
                      onClick={() => toggleAdmin(user.id, user.is_platform_admin)}
                      disabled={actionLoading === user.id}
                      className={`btn btn-small ${user.is_platform_admin ? 'btn-warning' : 'btn-success'}`}
                    >
                      {actionLoading === user.id ? '...' : user.is_platform_admin ? 'Revoke Admin' : 'Make Admin'}
                    </button>
                    <button
                      onClick={() => deleteUser(user.id)}
                      disabled={actionLoading === user.id || user.id === 'user-1'}
                      className="btn btn-small btn-danger"
                      title={user.id === 'user-1' ? 'Cannot delete yourself' : 'Delete user'}
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
