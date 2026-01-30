import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAdminAuth } from '../contexts/AdminAuthContext'

interface PlatformStats {
  total_users: number
  total_tenants: number
  total_workspaces: number
  total_documents: number
  total_projects: number
  admin_count: number
}

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8001'

export function DashboardPage() {
  const { getToken } = useAdminAuth()
  const [stats, setStats] = useState<PlatformStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    fetchStats()
  }, [])

  const fetchStats = async () => {
    try {
      const token = getToken()
      const response = await fetch(`${API_URL}/api/v1/admin/stats`, {
        headers: {
          'Authorization': `Bearer ${token}`,
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
          <Link to="/users" className="stat-link">Manage Users</Link>
        </div>

        <div className="stat-card">
          <div className="stat-value">{stats?.admin_count || 0}</div>
          <div className="stat-label">Platform Admins</div>
        </div>

        <div className="stat-card">
          <div className="stat-value">{stats?.total_tenants || 0}</div>
          <div className="stat-label">Total Tenants</div>
          <Link to="/tenants" className="stat-link">Manage Tenants</Link>
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
            <Link to="/users" className="btn btn-primary">
              Manage Users
            </Link>
            <Link to="/tenants" className="btn btn-secondary">
              Manage Tenants
            </Link>
          </div>
        </div>

        <div className="admin-section">
          <h2>Platform Features</h2>
          <ul style={{ paddingLeft: '1.5rem', color: 'var(--text-secondary)' }}>
            <li>User registration and authentication</li>
            <li>Social login (Google, GitHub, Microsoft, etc.)</li>
            <li>Enterprise SSO (OIDC, SAML, LDAP)</li>
            <li>Multi-factor authentication (MFA)</li>
            <li>Multi-tenancy with organizations</li>
            <li>Fine-grained authorization (OpenFGA)</li>
            <li>Role-based access control</li>
            <li>Password policies and account security</li>
          </ul>
        </div>

        <div className="admin-section">
          <h2>Architecture Overview</h2>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem' }}>
            This platform uses a headless authentication architecture with the following components:
          </p>
          <ul style={{ paddingLeft: '1.5rem', color: 'var(--text-secondary)' }}>
            <li><strong>API Gateway (Traefik)</strong> - Routes requests and handles ForwardAuth</li>
            <li><strong>AuthZ Service</strong> - JWT validation and permission checks</li>
            <li><strong>Identity Provider</strong> - User management, SSO, social login (internal)</li>
            <li><strong>OpenFGA</strong> - Fine-grained authorization (ReBAC/ABAC)</li>
            <li><strong>Sample API</strong> - Business logic backend</li>
          </ul>
        </div>
      </div>
    </div>
  )
}
