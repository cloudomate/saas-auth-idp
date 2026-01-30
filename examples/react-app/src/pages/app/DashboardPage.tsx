import { useSaasAuth, useTenant, useWorkspaces } from '@saas-starter/react'

export function DashboardPage() {
  const { user } = useSaasAuth()
  const { tenant, plan } = useTenant()
  const { workspaces, currentWorkspace } = useWorkspaces()

  return (
    <div>
      <h1 style={{ marginBottom: '2rem' }}>Dashboard</h1>

      {/* Stats */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-label">Current Plan</div>
          <div className="stat-value" style={{ fontSize: '1.5rem' }}>
            {plan?.name || 'Free'}
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Workspaces</div>
          <div className="stat-value">
            {workspaces.length}
            <span style={{ fontSize: '1rem', color: 'var(--text-secondary)' }}>
              {plan?.max_workspaces && plan.max_workspaces > 0
                ? ` / ${plan.max_workspaces}`
                : ''}
            </span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Current Workspace</div>
          <div className="stat-value" style={{ fontSize: '1.25rem' }}>
            {currentWorkspace?.display_name || '-'}
          </div>
        </div>
      </div>

      {/* Welcome card */}
      <div className="card" style={{ marginBottom: '1.5rem' }}>
        <h2 style={{ marginBottom: '1rem' }}>Welcome, {user?.name}!</h2>
        <p style={{ color: 'var(--text-secondary)' }}>
          You're logged in to <strong>{tenant?.display_name}</strong>. Use the
          sidebar to navigate between pages.
        </p>
      </div>

      {/* Quick info */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1.5rem' }}>
        <div className="card">
          <div className="card-header">
            <div className="card-title">Account Info</div>
          </div>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <tbody>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Email</td>
                <td style={{ padding: '0.5rem 0' }}>{user?.email}</td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Auth Provider</td>
                <td style={{ padding: '0.5rem 0', textTransform: 'capitalize' }}>
                  {user?.auth_provider}
                </td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Email Verified</td>
                <td style={{ padding: '0.5rem 0' }}>
                  {user?.email_verified ? (
                    <span style={{ color: 'var(--success)' }}>✓ Yes</span>
                  ) : (
                    <span style={{ color: 'var(--warning)' }}>✗ No</span>
                  )}
                </td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Role</td>
                <td style={{ padding: '0.5rem 0' }}>
                  {user?.is_platform_admin ? 'Platform Admin' : 'Member'}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div className="card">
          <div className="card-header">
            <div className="card-title">Organization Info</div>
          </div>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <tbody>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Name</td>
                <td style={{ padding: '0.5rem 0' }}>{tenant?.display_name}</td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Slug</td>
                <td style={{ padding: '0.5rem 0' }}>{tenant?.slug}</td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Plan</td>
                <td style={{ padding: '0.5rem 0' }}>{plan?.name}</td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Created</td>
                <td style={{ padding: '0.5rem 0' }}>
                  {tenant?.created_at
                    ? new Date(tenant.created_at).toLocaleDateString()
                    : '-'}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
