import { useState, useEffect, useCallback } from 'react'
import { useSaasAuth } from '@saas-starter/react'

interface Project {
  id: string
  name: string
  description: string
  workspace_id: string
  owner_id: string
  environment: 'production' | 'staging' | 'development'
  status: 'active' | 'paused' | 'archived'
  tags: string[]
  created_at: string
  updated_at: string
  permissions: {
    can_read: boolean
    can_write: boolean
    can_delete: boolean
    can_deploy: boolean
  }
}

interface Policy {
  name: string
  description: string
}

const SAMPLE_API_URL = 'http://localhost:8001'

const ENV_COLORS = {
  production: { bg: '#fef2f2', color: '#991b1b' },
  staging: { bg: '#fef9c3', color: '#854d0e' },
  development: { bg: '#dcfce7', color: '#166534' },
}

const STATUS_COLORS = {
  active: { bg: '#dcfce7', color: '#166534' },
  paused: { bg: '#fef9c3', color: '#854d0e' },
  archived: { bg: '#f3f4f6', color: '#6b7280' },
}

export function ProjectsPage() {
  const { user } = useSaasAuth()
  const [projects, setProjects] = useState<Project[]>([])
  const [policies, setPolicies] = useState<Policy[]>([])
  const [selectedProject, setSelectedProject] = useState<Project | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [deploying, setDeploying] = useState<string | null>(null)
  const [deployResult, setDeployResult] = useState<string | null>(null)

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)

  // Form states
  const [newName, setNewName] = useState('')
  const [newDescription, setNewDescription] = useState('')
  const [newEnvironment, setNewEnvironment] = useState<'production' | 'staging' | 'development'>('development')
  const [newTags, setNewTags] = useState('')

  // Admin toggle for demo
  const [isAdmin, setIsAdmin] = useState(false)

  const fetchProjects = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/projects?user_id=${user?.id}`, {
        headers: {
          'X-User-ID': user?.id || '',
          'X-Workspace-ID': 'workspace-1',
          'X-Is-Platform-Admin': isAdmin ? 'true' : 'false',
        },
      })
      const data = await res.json()
      setProjects(data.projects || [])
      setPolicies(data.policies || [])
    } catch {
      setError('Failed to fetch projects')
    } finally {
      setLoading(false)
    }
  }, [user?.id, isAdmin])

  useEffect(() => {
    fetchProjects()
  }, [fetchProjects])

  const fetchProject = async (id: string) => {
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/projects/${id}?user_id=${user?.id}`, {
        headers: {
          'X-User-ID': user?.id || '',
          'X-Workspace-ID': 'workspace-1',
          'X-Is-Platform-Admin': isAdmin ? 'true' : 'false',
        },
      })
      const data = await res.json()
      setSelectedProject(data.project)
    } catch {
      setError('Failed to fetch project')
    }
  }

  const createProject = async () => {
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/projects`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': user?.id || '',
          'X-Workspace-ID': 'workspace-1',
          'X-Is-Platform-Admin': isAdmin ? 'true' : 'false',
        },
        body: JSON.stringify({
          name: newName,
          description: newDescription,
          environment: newEnvironment,
          tags: newTags.split(',').map(t => t.trim()).filter(Boolean),
        }),
      })
      const data = await res.json()
      if (!res.ok) {
        throw new Error(data.message || 'Failed to create')
      }
      setShowCreateModal(false)
      setNewName('')
      setNewDescription('')
      setNewEnvironment('development')
      setNewTags('')
      fetchProjects()
    } catch (err: unknown) {
      const error = err as Error
      setError(error.message)
    }
  }

  const updateProject = async () => {
    if (!selectedProject) return
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/projects/${selectedProject.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': user?.id || '',
          'X-Workspace-ID': 'workspace-1',
          'X-Is-Platform-Admin': isAdmin ? 'true' : 'false',
        },
        body: JSON.stringify({
          name: newName,
          description: newDescription,
          environment: newEnvironment,
          tags: newTags.split(',').map(t => t.trim()).filter(Boolean),
        }),
      })
      const data = await res.json()
      if (!res.ok) {
        throw new Error(data.message || 'Failed to update')
      }
      setShowEditModal(false)
      fetchProjects()
      fetchProject(selectedProject.id)
    } catch (err: unknown) {
      const error = err as Error
      setError(error.message)
    }
  }

  const deleteProject = async (id: string) => {
    if (!confirm('Are you sure you want to delete this project?')) return
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/projects/${id}`, {
        method: 'DELETE',
        headers: {
          'X-User-ID': user?.id || '',
          'X-Workspace-ID': 'workspace-1',
          'X-Is-Platform-Admin': isAdmin ? 'true' : 'false',
        },
      })
      const data = await res.json()
      if (!res.ok) {
        throw new Error(data.message || 'Failed to delete')
      }
      setSelectedProject(null)
      fetchProjects()
    } catch (err: unknown) {
      const error = err as Error
      setError(error.message)
    }
  }

  const deployProject = async (id: string) => {
    setDeploying(id)
    setDeployResult(null)
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/projects/${id}/deploy`, {
        method: 'POST',
        headers: {
          'X-User-ID': user?.id || '',
          'X-Workspace-ID': 'workspace-1',
          'X-Is-Platform-Admin': isAdmin ? 'true' : 'false',
        },
      })
      const data = await res.json()
      if (!res.ok) {
        throw new Error(data.message || 'Deployment failed')
      }
      setDeployResult(`Deployed successfully at ${new Date(data.deployed_at).toLocaleTimeString()}`)
    } catch (err: unknown) {
      const error = err as Error
      setDeployResult(`Error: ${error.message}`)
    } finally {
      setDeploying(null)
    }
  }

  const openEditModal = (proj: Project) => {
    setNewName(proj.name)
    setNewDescription(proj.description)
    setNewEnvironment(proj.environment)
    setNewTags(proj.tags.join(', '))
    setShowEditModal(true)
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <div>
          <h1>Projects</h1>
          <p style={{ color: 'var(--text-secondary)' }}>
            ABAC Demo: Access based on attributes (environment, status, role)
          </p>
        </div>
        <button className="btn btn-primary" style={{ width: 'auto' }} onClick={() => setShowCreateModal(true)}>
          + New Project
        </button>
      </div>

      {/* Admin Toggle for Demo */}
      <div className="card" style={{ marginBottom: '1.5rem', padding: '1rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <strong>Demo Mode</strong>
            <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', margin: 0 }}>
              Toggle admin role to see how permissions change
            </p>
          </div>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
            <input
              type="checkbox"
              checked={isAdmin}
              onChange={e => setIsAdmin(e.target.checked)}
              style={{ width: 20, height: 20 }}
            />
            <span>{isAdmin ? '👑 Admin' : '👤 Member'}</span>
          </label>
        </div>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {deployResult && (
        <div className={`alert ${deployResult.startsWith('Error') ? 'alert-error' : 'alert-success'}`}>
          {deployResult}
        </div>
      )}

      {/* ABAC Policies Info */}
      <div className="card" style={{ marginBottom: '1.5rem' }}>
        <div className="card-header">
          <div className="card-title">Active ABAC Policies</div>
        </div>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '0.5rem' }}>
          {policies.map((policy, i) => (
            <div
              key={i}
              style={{
                padding: '0.5rem 0.75rem',
                background: 'var(--background)',
                borderRadius: 'var(--radius)',
                fontSize: '0.875rem',
              }}
              title={policy.description}
            >
              📋 {policy.description}
            </div>
          ))}
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
        {/* Project List */}
        <div className="card">
          <div className="card-header">
            <div className="card-title">All Projects</div>
          </div>

          {loading ? (
            <div className="loading"><div className="spinner" /></div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {projects.map(proj => (
                <div
                  key={proj.id}
                  onClick={() => fetchProject(proj.id)}
                  style={{
                    padding: '1rem',
                    background: selectedProject?.id === proj.id ? 'var(--background)' : 'transparent',
                    borderRadius: 'var(--radius)',
                    cursor: 'pointer',
                    border: '1px solid var(--border)',
                  }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                    <div>
                      <div style={{ fontWeight: 500 }}>{proj.name}</div>
                      <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.25rem' }}>
                        <span
                          style={{
                            fontSize: '0.75rem',
                            padding: '0.125rem 0.5rem',
                            borderRadius: '4px',
                            ...ENV_COLORS[proj.environment],
                          }}
                        >
                          {proj.environment}
                        </span>
                        <span
                          style={{
                            fontSize: '0.75rem',
                            padding: '0.125rem 0.5rem',
                            borderRadius: '4px',
                            ...STATUS_COLORS[proj.status],
                          }}
                        >
                          {proj.status}
                        </span>
                      </div>
                    </div>
                    <div style={{ display: 'flex', gap: '0.25rem' }}>
                      {proj.permissions.can_write && (
                        <span title="Can edit" style={{ opacity: 1 }}>✏️</span>
                      )}
                      {proj.permissions.can_deploy && (
                        <span title="Can deploy" style={{ opacity: 1 }}>🚀</span>
                      )}
                      {proj.permissions.can_delete && (
                        <span title="Can delete" style={{ opacity: 1 }}>🗑️</span>
                      )}
                      {!proj.permissions.can_write && !proj.permissions.can_deploy && (
                        <span title="Read only" style={{ opacity: 0.5 }}>👁️</span>
                      )}
                    </div>
                  </div>
                </div>
              ))}
              {projects.length === 0 && (
                <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: '2rem' }}>
                  No projects yet
                </p>
              )}
            </div>
          )}
        </div>

        {/* Project Detail */}
        <div className="card">
          <div className="card-header">
            <div className="card-title">
              {selectedProject ? selectedProject.name : 'Select a project'}
            </div>
            {selectedProject && (
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                {selectedProject.permissions.can_write && (
                  <button className="btn btn-secondary btn-small" onClick={() => openEditModal(selectedProject)}>
                    Edit
                  </button>
                )}
                {selectedProject.permissions.can_deploy && (
                  <button
                    className="btn btn-primary btn-small"
                    onClick={() => deployProject(selectedProject.id)}
                    disabled={deploying === selectedProject.id}
                  >
                    {deploying === selectedProject.id ? '🚀 Deploying...' : '🚀 Deploy'}
                  </button>
                )}
                {selectedProject.permissions.can_delete && (
                  <button className="btn btn-danger btn-small" onClick={() => deleteProject(selectedProject.id)}>
                    Delete
                  </button>
                )}
              </div>
            )}
          </div>

          {selectedProject ? (
            <div>
              <div style={{ marginBottom: '1rem' }}>
                <p style={{ color: 'var(--text-secondary)' }}>{selectedProject.description}</p>
                <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.5rem' }}>
                  <span style={{ fontSize: '0.875rem', padding: '0.25rem 0.75rem', borderRadius: '4px', ...ENV_COLORS[selectedProject.environment] }}>
                    {selectedProject.environment}
                  </span>
                  <span style={{ fontSize: '0.875rem', padding: '0.25rem 0.75rem', borderRadius: '4px', ...STATUS_COLORS[selectedProject.status] }}>
                    {selectedProject.status}
                  </span>
                </div>
              </div>

              {/* Tags */}
              {selectedProject.tags.length > 0 && (
                <div style={{ marginBottom: '1rem' }}>
                  <h4 style={{ marginBottom: '0.5rem' }}>Tags</h4>
                  <div style={{ display: 'flex', gap: '0.25rem', flexWrap: 'wrap' }}>
                    {selectedProject.tags.map((tag, i) => (
                      <span key={i} style={{ padding: '0.25rem 0.5rem', background: 'var(--background)', borderRadius: '4px', fontSize: '0.75rem' }}>
                        #{tag}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Permissions */}
              <div style={{ marginBottom: '1rem' }}>
                <h4 style={{ marginBottom: '0.5rem' }}>Your Permissions (ABAC Evaluated)</h4>
                <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                  {Object.entries(selectedProject.permissions).map(([perm, allowed]) => (
                    <span
                      key={perm}
                      style={{
                        padding: '0.25rem 0.5rem',
                        borderRadius: '4px',
                        fontSize: '0.75rem',
                        background: allowed ? '#dcfce7' : '#fee2e2',
                        color: allowed ? '#166534' : '#991b1b',
                      }}
                    >
                      {allowed ? '✓' : '✗'} {perm.replace('can_', '')}
                    </span>
                  ))}
                </div>
              </div>

              {/* Why permissions */}
              <div style={{ padding: '1rem', background: 'var(--background)', borderRadius: 'var(--radius)', fontSize: '0.875rem' }}>
                <strong>Why these permissions?</strong>
                <ul style={{ margin: '0.5rem 0 0 1.25rem', padding: 0 }}>
                  {selectedProject.environment === 'production' && !isAdmin && (
                    <li style={{ color: 'var(--danger)' }}>Production projects require admin role for write/deploy</li>
                  )}
                  {selectedProject.status === 'archived' && (
                    <li style={{ color: 'var(--danger)' }}>Archived projects are read-only</li>
                  )}
                  {selectedProject.status === 'paused' && (
                    <li style={{ color: 'var(--warning)' }}>Paused projects cannot be deployed</li>
                  )}
                  {isAdmin && (
                    <li style={{ color: 'var(--success)' }}>Admin role grants elevated permissions</li>
                  )}
                  {!isAdmin && selectedProject.owner_id === user?.id && (
                    <li style={{ color: 'var(--primary)' }}>You are the owner of this project</li>
                  )}
                </ul>
              </div>
            </div>
          ) : (
            <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: '2rem' }}>
              Select a project to view details
            </p>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <div className="modal-title">Create Project</div>
              <button className="modal-close" onClick={() => setShowCreateModal(false)}>×</button>
            </div>
            <div className="form-group">
              <label>Name</label>
              <input value={newName} onChange={e => setNewName(e.target.value)} placeholder="Project name" />
            </div>
            <div className="form-group">
              <label>Description</label>
              <textarea
                value={newDescription}
                onChange={e => setNewDescription(e.target.value)}
                placeholder="Project description..."
                rows={3}
                style={{ width: '100%', padding: '0.75rem', border: '1px solid var(--border)', borderRadius: 'var(--radius)' }}
              />
            </div>
            <div className="form-group">
              <label>Environment</label>
              <select value={newEnvironment} onChange={e => setNewEnvironment(e.target.value as typeof newEnvironment)} style={{ width: '100%', padding: '0.75rem', border: '1px solid var(--border)', borderRadius: 'var(--radius)' }}>
                <option value="development">Development</option>
                <option value="staging">Staging</option>
                <option value="production">Production {!isAdmin && '(requires admin)'}</option>
              </select>
              {newEnvironment === 'production' && !isAdmin && (
                <small style={{ color: 'var(--danger)' }}>⚠️ You need admin role to create production projects</small>
              )}
            </div>
            <div className="form-group">
              <label>Tags (comma separated)</label>
              <input value={newTags} onChange={e => setNewTags(e.target.value)} placeholder="backend, api, go" />
            </div>
            <div style={{ display: 'flex', gap: '1rem' }}>
              <button className="btn btn-secondary" onClick={() => setShowCreateModal(false)}>Cancel</button>
              <button className="btn btn-primary" onClick={createProject}>Create</button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Modal */}
      {showEditModal && selectedProject && (
        <div className="modal-overlay" onClick={() => setShowEditModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <div className="modal-title">Edit Project</div>
              <button className="modal-close" onClick={() => setShowEditModal(false)}>×</button>
            </div>
            <div className="form-group">
              <label>Name</label>
              <input value={newName} onChange={e => setNewName(e.target.value)} />
            </div>
            <div className="form-group">
              <label>Description</label>
              <textarea
                value={newDescription}
                onChange={e => setNewDescription(e.target.value)}
                rows={3}
                style={{ width: '100%', padding: '0.75rem', border: '1px solid var(--border)', borderRadius: 'var(--radius)' }}
              />
            </div>
            <div className="form-group">
              <label>Environment</label>
              <select
                value={newEnvironment}
                onChange={e => setNewEnvironment(e.target.value as typeof newEnvironment)}
                style={{ width: '100%', padding: '0.75rem', border: '1px solid var(--border)', borderRadius: 'var(--radius)' }}
              >
                <option value="development">Development</option>
                <option value="staging">Staging</option>
                <option value="production">Production</option>
              </select>
              {(newEnvironment === 'production' || selectedProject.environment === 'production') && !isAdmin && (
                <small style={{ color: 'var(--danger)' }}>⚠️ Changing to/from production requires admin</small>
              )}
            </div>
            <div className="form-group">
              <label>Tags</label>
              <input value={newTags} onChange={e => setNewTags(e.target.value)} />
            </div>
            <div style={{ display: 'flex', gap: '1rem' }}>
              <button className="btn btn-secondary" onClick={() => setShowEditModal(false)}>Cancel</button>
              <button className="btn btn-primary" onClick={updateProject}>Save</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
