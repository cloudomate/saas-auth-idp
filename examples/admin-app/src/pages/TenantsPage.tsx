import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'

interface Tenant {
  id: string
  name: string
  slug: string
  plan: string
  owner_id: string
  created_at: string
  updated_at: string
}

interface Workspace {
  id: string
  name: string
  tenant_id: string
  created_at: string
}

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8001'

export function TenantsPage() {
  const [tenants, setTenants] = useState<Tenant[]>([])
  const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null)
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  useEffect(() => {
    fetchTenants()
  }, [])

  const fetchTenants = async () => {
    try {
      const response = await fetch(`${API_URL}/api/v1/admin/tenants`, {
        headers: {
          'X-User-ID': 'user-1',
          'X-Is-Platform-Admin': 'true',
        },
      })

      if (!response.ok) {
        throw new Error('Failed to fetch tenants')
      }

      const data = await response.json()
      setTenants(data.tenants || [])
    } catch (err) {
      setError('Failed to load tenants')
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const viewTenantDetails = async (tenantId: string) => {
    setActionLoading(tenantId)
    try {
      const response = await fetch(`${API_URL}/api/v1/admin/tenants/${tenantId}`, {
        headers: {
          'X-User-ID': 'user-1',
          'X-Is-Platform-Admin': 'true',
        },
      })

      if (!response.ok) {
        throw new Error('Failed to fetch tenant details')
      }

      const data = await response.json()
      setSelectedTenant(data.tenant)
      setWorkspaces(data.workspaces || [])
    } catch (err) {
      setError('Failed to load tenant details')
      console.error(err)
    } finally {
      setActionLoading(null)
    }
  }

  const deleteTenant = async (tenantId: string) => {
    if (!confirm('Are you sure you want to delete this tenant? This will also delete all associated workspaces.')) {
      return
    }

    setActionLoading(tenantId)
    try {
      const response = await fetch(`${API_URL}/api/v1/admin/tenants/${tenantId}`, {
        method: 'DELETE',
        headers: {
          'X-User-ID': 'user-1',
          'X-Is-Platform-Admin': 'true',
        },
      })

      if (!response.ok) {
        throw new Error('Failed to delete tenant')
      }

      if (selectedTenant?.id === tenantId) {
        setSelectedTenant(null)
        setWorkspaces([])
      }

      await fetchTenants()
    } catch (err) {
      setError('Failed to delete tenant')
      console.error(err)
    } finally {
      setActionLoading(null)
    }
  }

  const deleteWorkspace = async (workspaceId: string) => {
    if (!confirm('Are you sure you want to delete this workspace?')) {
      return
    }

    setActionLoading(workspaceId)
    try {
      const response = await fetch(`${API_URL}/api/v1/admin/workspaces/${workspaceId}`, {
        method: 'DELETE',
        headers: {
          'X-User-ID': 'user-1',
          'X-Is-Platform-Admin': 'true',
        },
      })

      if (!response.ok) {
        throw new Error('Failed to delete workspace')
      }

      if (selectedTenant) {
        await viewTenantDetails(selectedTenant.id)
      }
    } catch (err) {
      setError('Failed to delete workspace')
      console.error(err)
    } finally {
      setActionLoading(null)
    }
  }

  if (loading) {
    return (
      <div className="admin-container">
        <div className="loading">Loading tenants...</div>
      </div>
    )
  }

  return (
    <div className="admin-container">
      <div className="admin-header">
        <div className="header-with-back">
          <Link to="/" className="back-link">&larr; Back to Dashboard</Link>
          <h1>Tenant Management</h1>
        </div>
        <p>Manage organizations and their workspaces</p>
      </div>

      {error && (
        <div className="alert alert-error" style={{ marginBottom: '1rem' }}>
          {error}
          <button onClick={() => setError('')} className="alert-close">&times;</button>
        </div>
      )}

      <div className="admin-grid">
        <div className="admin-panel">
          <h2>Tenants ({tenants.length})</h2>
          <div className="admin-table-container">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Slug</th>
                  <th>Plan</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {tenants.map((tenant) => (
                  <tr
                    key={tenant.id}
                    className={selectedTenant?.id === tenant.id ? 'selected' : ''}
                  >
                    <td>{tenant.name}</td>
                    <td><code>{tenant.slug}</code></td>
                    <td>
                      <span className={`badge badge-plan-${tenant.plan}`}>
                        {tenant.plan}
                      </span>
                    </td>
                    <td>
                      <div className="action-buttons-small">
                        <button
                          onClick={() => viewTenantDetails(tenant.id)}
                          disabled={actionLoading === tenant.id}
                          className="btn btn-small btn-secondary"
                        >
                          {actionLoading === tenant.id ? '...' : 'View'}
                        </button>
                        <button
                          onClick={() => deleteTenant(tenant.id)}
                          disabled={actionLoading === tenant.id}
                          className="btn btn-small btn-danger"
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

        {selectedTenant && (
          <div className="admin-panel">
            <h2>Tenant Details: {selectedTenant.name}</h2>
            <div className="detail-grid">
              <div className="detail-item">
                <label>ID</label>
                <span><code>{selectedTenant.id}</code></span>
              </div>
              <div className="detail-item">
                <label>Slug</label>
                <span><code>{selectedTenant.slug}</code></span>
              </div>
              <div className="detail-item">
                <label>Plan</label>
                <span className={`badge badge-plan-${selectedTenant.plan}`}>
                  {selectedTenant.plan}
                </span>
              </div>
              <div className="detail-item">
                <label>Owner ID</label>
                <span><code>{selectedTenant.owner_id}</code></span>
              </div>
              <div className="detail-item">
                <label>Created</label>
                <span>{new Date(selectedTenant.created_at).toLocaleString()}</span>
              </div>
            </div>

            <h3>Workspaces ({workspaces.length})</h3>
            {workspaces.length === 0 ? (
              <p className="no-data">No workspaces found</p>
            ) : (
              <div className="workspace-list">
                {workspaces.map((workspace) => (
                  <div key={workspace.id} className="workspace-item">
                    <div className="workspace-info">
                      <strong>{workspace.name}</strong>
                      <code>{workspace.id}</code>
                    </div>
                    <button
                      onClick={() => deleteWorkspace(workspace.id)}
                      disabled={actionLoading === workspace.id}
                      className="btn btn-small btn-danger"
                    >
                      {actionLoading === workspace.id ? '...' : 'Delete'}
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
