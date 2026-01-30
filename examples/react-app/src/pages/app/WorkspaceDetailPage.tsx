import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useWorkspaces, useHierarchy } from '@saas-starter/react'
import type { Membership } from '@saas-starter/react'

export function WorkspaceDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { workspaces, setCurrentWorkspace } = useWorkspaces()
  const { listMembers, addMember } = useHierarchy()

  const [members, setMembers] = useState<Membership[]>([])
  const [loadingMembers, setLoadingMembers] = useState(true)
  const [showAddMember, setShowAddMember] = useState(false)
  const [newMemberEmail, setNewMemberEmail] = useState('')
  const [newMemberRole, setNewMemberRole] = useState('member')
  const [addingMember, setAddingMember] = useState(false)
  const [error, setError] = useState('')

  const workspace = workspaces.find((w) => w.id === id || w.slug === id)

  useEffect(() => {
    if (workspace) {
      setCurrentWorkspace(workspace.id)
      loadMembers()
    }
  }, [workspace?.id])

  const loadMembers = async () => {
    if (!workspace) return
    setLoadingMembers(true)
    try {
      const membersList = await listMembers('workspace', workspace.id)
      setMembers(membersList)
    } catch (err) {
      console.error('Failed to load members:', err)
    } finally {
      setLoadingMembers(false)
    }
  }

  const handleAddMember = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!workspace) return

    setError('')
    setAddingMember(true)

    try {
      await addMember('workspace', workspace.id, newMemberEmail, newMemberRole)
      setShowAddMember(false)
      setNewMemberEmail('')
      setNewMemberRole('member')
      await loadMembers()
    } catch (err: unknown) {
      const error = err as { message?: string }
      setError(error.message || 'Failed to add member')
    } finally {
      setAddingMember(false)
    }
  }

  if (!workspace) {
    return (
      <div>
        <p>Workspace not found</p>
        <button className="btn btn-secondary" onClick={() => navigate('/workspaces')}>
          Back to Workspaces
        </button>
      </div>
    )
  }

  return (
    <div>
      <div style={{ marginBottom: '2rem' }}>
        <button
          onClick={() => navigate('/workspaces')}
          style={{
            background: 'none',
            border: 'none',
            color: 'var(--primary)',
            cursor: 'pointer',
            marginBottom: '0.5rem',
          }}
        >
          ← Back to Workspaces
        </button>
        <h1>{workspace.display_name}</h1>
        <p style={{ color: 'var(--text-secondary)' }}>/{workspace.slug}</p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '2rem' }}>
        {/* Members section */}
        <div className="card">
          <div className="card-header">
            <div className="card-title">Members</div>
            <button
              className="btn btn-primary btn-small"
              style={{ width: 'auto' }}
              onClick={() => setShowAddMember(true)}
            >
              + Add Member
            </button>
          </div>

          {loadingMembers ? (
            <div className="loading">
              <div className="spinner" />
            </div>
          ) : (
            <div className="members-list">
              {members.map((member) => (
                <div key={member.user_id} className="member-item">
                  <div className="member-info">
                    <div className="user-avatar" style={{ width: 36, height: 36 }}>
                      {member.picture ? (
                        <img src={member.picture} alt={member.name} />
                      ) : (
                        member.name.charAt(0).toUpperCase()
                      )}
                    </div>
                    <div>
                      <div style={{ fontWeight: 500 }}>{member.name}</div>
                      <div style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
                        {member.email}
                      </div>
                    </div>
                  </div>
                  <span className="member-role">{member.role}</span>
                </div>
              ))}

              {members.length === 0 && (
                <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: '1rem' }}>
                  No members yet
                </p>
              )}
            </div>
          )}
        </div>

        {/* Workspace info */}
        <div className="card">
          <div className="card-header">
            <div className="card-title">Details</div>
          </div>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <tbody>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>ID</td>
                <td style={{ padding: '0.5rem 0', fontSize: '0.75rem', fontFamily: 'monospace' }}>
                  {workspace.id}
                </td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Slug</td>
                <td style={{ padding: '0.5rem 0' }}>{workspace.slug}</td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Created</td>
                <td style={{ padding: '0.5rem 0' }}>
                  {new Date(workspace.created_at).toLocaleDateString()}
                </td>
              </tr>
              <tr>
                <td style={{ padding: '0.5rem 0', color: 'var(--text-secondary)' }}>Members</td>
                <td style={{ padding: '0.5rem 0' }}>{members.length}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Add Member Modal */}
      {showAddMember && (
        <div className="modal-overlay" onClick={() => setShowAddMember(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div className="modal-title">Add Member</div>
              <button className="modal-close" onClick={() => setShowAddMember(false)}>
                ×
              </button>
            </div>

            {error && <div className="alert alert-error">{error}</div>}

            <form onSubmit={handleAddMember}>
              <div className="form-group">
                <label htmlFor="member-email">Email</label>
                <input
                  id="member-email"
                  type="email"
                  value={newMemberEmail}
                  onChange={(e) => setNewMemberEmail(e.target.value)}
                  placeholder="member@example.com"
                  required
                />
              </div>
              <div className="form-group">
                <label htmlFor="member-role">Role</label>
                <select
                  id="member-role"
                  value={newMemberRole}
                  onChange={(e) => setNewMemberRole(e.target.value)}
                  style={{
                    width: '100%',
                    padding: '0.75rem',
                    border: '1px solid var(--border)',
                    borderRadius: 'var(--radius)',
                  }}
                >
                  <option value="viewer">Viewer (read-only)</option>
                  <option value="member">Member (read/write)</option>
                  <option value="admin">Admin (full access)</option>
                </select>
              </div>
              <div style={{ display: 'flex', gap: '1rem' }}>
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => setShowAddMember(false)}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={addingMember}
                >
                  {addingMember ? 'Adding...' : 'Add Member'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
