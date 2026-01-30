import { useSaasAuth, useTenant } from '@saas-starter/react'

export function SettingsPage() {
  const { user, refreshUser } = useSaasAuth()
  const { tenant, plan, subscription, refreshTenant } = useTenant()

  return (
    <div>
      <h1 style={{ marginBottom: '2rem' }}>Settings</h1>

      <div style={{ display: 'grid', gap: '1.5rem', maxWidth: '800px' }}>
        {/* Profile Settings */}
        <div className="card">
          <div className="card-header">
            <div className="card-title">Profile</div>
            <button
              className="btn btn-secondary btn-small"
              style={{ width: 'auto' }}
              onClick={refreshUser}
            >
              Refresh
            </button>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '1rem', alignItems: 'center' }}>
            <div
              className="user-avatar"
              style={{ width: 64, height: 64, fontSize: '1.5rem' }}
            >
              {user?.picture ? (
                <img src={user.picture} alt={user?.name} />
              ) : (
                user?.name?.charAt(0).toUpperCase()
              )}
            </div>
            <div>
              <div style={{ fontWeight: 600, fontSize: '1.25rem' }}>{user?.name}</div>
              <div style={{ color: 'var(--text-secondary)' }}>{user?.email}</div>
            </div>
          </div>

          <div style={{ marginTop: '1.5rem', paddingTop: '1.5rem', borderTop: '1px solid var(--border)' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <tbody>
                <tr>
                  <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)', width: '150px' }}>
                    User ID
                  </td>
                  <td style={{ padding: '0.5rem 0', fontFamily: 'monospace', fontSize: '0.875rem' }}>
                    {user?.id}
                  </td>
                </tr>
                <tr>
                  <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>
                    Auth Provider
                  </td>
                  <td style={{ padding: '0.5rem 0', textTransform: 'capitalize' }}>
                    {user?.auth_provider}
                  </td>
                </tr>
                <tr>
                  <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>
                    Email Status
                  </td>
                  <td style={{ padding: '0.5rem 0' }}>
                    {user?.email_verified ? (
                      <span style={{ color: 'var(--success)' }}>Verified</span>
                    ) : (
                      <span style={{ color: 'var(--warning)' }}>Not verified</span>
                    )}
                  </td>
                </tr>
                <tr>
                  <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>
                    Account Created
                  </td>
                  <td style={{ padding: '0.5rem 0' }}>
                    {user?.created_at
                      ? new Date(user.created_at).toLocaleDateString()
                      : '-'}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        {/* Organization Settings */}
        <div className="card">
          <div className="card-header">
            <div className="card-title">Organization</div>
            <button
              className="btn btn-secondary btn-small"
              style={{ width: 'auto' }}
              onClick={refreshTenant}
            >
              Refresh
            </button>
          </div>

          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <tbody>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)', width: '150px' }}>
                  Name
                </td>
                <td style={{ padding: '0.5rem 0' }}>{tenant?.display_name}</td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>
                  Slug
                </td>
                <td style={{ padding: '0.5rem 0' }}>{tenant?.slug}</td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>
                  Tenant ID
                </td>
                <td style={{ padding: '0.5rem 0', fontFamily: 'monospace', fontSize: '0.875rem' }}>
                  {tenant?.id}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        {/* Subscription Settings */}
        <div className="card">
          <div className="card-header">
            <div className="card-title">Subscription</div>
          </div>

          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '1rem',
              background: 'var(--background)',
              borderRadius: 'var(--radius)',
              marginBottom: '1rem',
            }}
          >
            <div>
              <div style={{ fontWeight: 600, fontSize: '1.25rem' }}>{plan?.name}</div>
              <div style={{ color: 'var(--text-secondary)' }}>{plan?.description}</div>
            </div>
            <div style={{ textAlign: 'right' }}>
              <div style={{ fontSize: '1.5rem', fontWeight: 700, color: 'var(--primary)' }}>
                ${plan?.monthly_price}
                <span style={{ fontSize: '1rem', color: 'var(--text-secondary)' }}>/mo</span>
              </div>
            </div>
          </div>

          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <tbody>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)', width: '150px' }}>
                  Status
                </td>
                <td style={{ padding: '0.5rem 0' }}>
                  <span
                    style={{
                      display: 'inline-block',
                      padding: '0.25rem 0.75rem',
                      borderRadius: '9999px',
                      fontSize: '0.875rem',
                      background:
                        subscription?.status === 'active'
                          ? '#dcfce7'
                          : subscription?.status === 'trialing'
                          ? '#dbeafe'
                          : '#fee2e2',
                      color:
                        subscription?.status === 'active'
                          ? '#166534'
                          : subscription?.status === 'trialing'
                          ? '#1e40af'
                          : '#991b1b',
                    }}
                  >
                    {subscription?.status || 'Active'}
                  </span>
                </td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>
                  Max Workspaces
                </td>
                <td style={{ padding: '0.5rem 0' }}>
                  {plan?.max_workspaces && plan.max_workspaces > 0
                    ? plan.max_workspaces
                    : 'Unlimited'}
                </td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>
                  Max Users
                </td>
                <td style={{ padding: '0.5rem 0' }}>
                  {plan?.max_users && plan.max_users > 0 ? plan.max_users : 'Unlimited'}
                </td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>
                  Features
                </td>
                <td style={{ padding: '0.5rem 0' }}>
                  <ul style={{ margin: 0, paddingLeft: '1.25rem' }}>
                    {plan?.features?.map((feature, i) => (
                      <li key={i} style={{ padding: '0.125rem 0' }}>
                        {feature}
                      </li>
                    ))}
                  </ul>
                </td>
              </tr>
            </tbody>
          </table>

          <div style={{ marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid var(--border)' }}>
            <button className="btn btn-primary" disabled>
              Upgrade Plan (Coming Soon)
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
