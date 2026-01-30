import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'

interface PlatformStats {
  total_users: number
  total_tenants: number
  total_workspaces: number
  total_documents: number
  total_projects: number
  admin_count: number
}

export function AdminDashboardPage() {
  const [stats, setStats] = useState<PlatformStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    fetchStats()
  }, [])

  const fetchStats = async () => {
    try {
      const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8001'
      const response = await fetch(`${apiUrl}/api/v1/admin/stats`, {
        headers: {
          'X-User-ID': 'user-1',
          'X-Is-Platform-Admin': 'true',
        },
      })

      if (!response.ok) {
        throw new Error('Failed to fetch stats')
      }

      const data = await response.json()
      setStats(data)
    } catch (err) {
      setError('Failed to load platform statistics')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="admin-container">
        <div className="loading">Loading platform statistics...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="admin-container">
        <div className="alert alert-error">{error}</div>
      </div>
    )
  }

  return (
    <div className="admin-container">
      <div className="admin-header">
        <h1>Platform Admin Dashboard</h1>
        <p>Manage users, tenants, and platform resources</p>
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-value">{stats?.total_users || 0}</div>
          <div className="stat-label">Total Users</div>
          <Link to="/admin/users" className="stat-link">Manage Users</Link>
        </div>

        <div className="stat-card">
          <div className="stat-value">{stats?.admin_count || 0}</div>
          <div className="stat-label">Platform Admins</div>
        </div>

        <div className="stat-card">
          <div className="stat-value">{stats?.total_tenants || 0}</div>
          <div className="stat-label">Total Tenants</div>
          <Link to="/admin/tenants" className="stat-link">Manage Tenants</Link>
        </div>

        <div className="stat-card">
          <div className="stat-value">{stats?.total_workspaces || 0}</div>
          <div className="stat-label">Total Workspaces</div>
        </div>

        <div className="stat-card">
          <div className="stat-value">{stats?.total_documents || 0}</div>
          <div className="stat-label">Total Documents</div>
        </div>

        <div className="stat-card">
          <div className="stat-value">{stats?.total_projects || 0}</div>
          <div className="stat-label">Total Projects</div>
        </div>
      </div>

      <div className="admin-sections">
        <div className="admin-section">
          <h2>Quick Actions</h2>
          <div className="action-buttons">
            <Link to="/admin/users" className="btn btn-primary">
              Manage Users
            </Link>
            <Link to="/admin/tenants" className="btn btn-secondary">
              Manage Tenants
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}
