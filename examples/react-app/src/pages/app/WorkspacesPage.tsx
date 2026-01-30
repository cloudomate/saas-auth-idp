import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useWorkspaces, useTenant } from '@saas-starter/react'

export function WorkspacesPage() {
  const navigate = useNavigate()
  const { workspaces, createWorkspace, refreshWorkspaces, isLoading, error } = useWorkspaces()
  const { plan } = useTenant()

  const [showModal, setShowModal] = useState(false)
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [createError, setCreateError] = useState('')
  const [creating, setCreating] = useState(false)

  const canCreateMore =
    !plan?.max_workspaces ||
    plan.max_workspaces < 0 ||
    workspaces.length < plan.max_workspaces

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setCreateError('')
    setCreating(true)

    try {
      const workspace = await createWorkspace(name, slug)
      setShowModal(false)
      setName('')
      setSlug('')
      await refreshWorkspaces()
      navigate(`/workspaces/${workspace.id}`)
    } catch (err: unknown) {
      const error = err as { message?: string }
      setCreateError(error.message || 'Failed to create workspace')
    } finally {
      setCreating(false)
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h1>Workspaces</h1>
        <button
          className="btn btn-primary"
          style={{ width: 'auto' }}
          onClick={() => setShowModal(true)}
          disabled={!canCreateMore}
        >
          + New Workspace
        </button>
      </div>

      {!canCreateMore && (
        <div className="alert alert-warning" style={{ background: '#fffbeb', borderColor: '#fcd34d', color: '#b45309' }}>
          You've reached the workspace limit for your plan ({plan?.max_workspaces} workspaces).
          Upgrade to create more.
        </div>
      )}

      {error && (
        <div className="alert alert-error">{error.message}</div>
      )}

      {isLoading ? (
        <div className="loading">
          <div className="spinner" />
        </div>
      ) : (
        <div className="workspaces-grid">
          {workspaces.map((workspace) => (
            <div
              key={workspace.id}
              className="workspace-card"
              onClick={() => navigate(`/workspaces/${workspace.id}`)}
            >
              <div className="workspace-name">{workspace.display_name}</div>
              <div className="workspace-slug">/{workspace.slug}</div>
              <div className="workspace-meta">
                <span>
                  Created {new Date(workspace.created_at).toLocaleDateString()}
                </span>
              </div>
            </div>
          ))}

          {workspaces.length === 0 && (
            <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: '3rem', color: 'var(--text-secondary)' }}>
              <p>No workspaces yet. Create your first workspace to get started.</p>
            </div>
          )}
        </div>
      )}

      {/* Create Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <div className="modal-title">Create Workspace</div>
              <button className="modal-close" onClick={() => setShowModal(false)}>
                ×
              </button>
            </div>

            {createError && (
              <div className="alert alert-error">{createError}</div>
            )}

            <form onSubmit={handleCreate}>
              <div className="form-group">
                <label htmlFor="ws-name">Workspace Name</label>
                <input
                  id="ws-name"
                  type="text"
                  value={name}
                  onChange={(e) => {
                    setName(e.target.value)
                    if (!slug) {
                      setSlug(
                        e.target.value
                          .toLowerCase()
                          .replace(/[^a-z0-9]+/g, '-')
                          .replace(/^-|-$/g, '')
                      )
                    }
                  }}
                  placeholder="Engineering"
                  required
                />
              </div>
              <div className="form-group">
                <label htmlFor="ws-slug">Slug</label>
                <input
                  id="ws-slug"
                  type="text"
                  value={slug}
                  onChange={(e) =>
                    setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))
                  }
                  placeholder="engineering"
                  required
                />
              </div>
              <div style={{ display: 'flex', gap: '1rem' }}>
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => setShowModal(false)}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={creating}
                >
                  {creating ? 'Creating...' : 'Create Workspace'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
