import { useState, useEffect, useCallback } from 'react'
import { useSaasAuth } from '@saas-starter/react'

interface Document {
  id: string
  title: string
  content: string
  workspace_id: string
  owner_id: string
  visibility: 'public' | 'workspace' | 'private'
  status: 'draft' | 'published' | 'archived'
  created_at: string
  updated_at: string
  permissions: {
    can_read: boolean
    can_write: boolean
    can_delete: boolean
    can_share: boolean
  }
}

interface Share {
  document_id: string
  user_id: string
  role: string
}

const SAMPLE_API_URL = 'http://localhost:8001'

export function DocumentsPage() {
  const { user } = useSaasAuth()
  const [documents, setDocuments] = useState<Document[]>([])
  const [selectedDoc, setSelectedDoc] = useState<Document | null>(null)
  const [shares, setShares] = useState<Share[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showShareModal, setShowShareModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)

  // Form states
  const [newTitle, setNewTitle] = useState('')
  const [newContent, setNewContent] = useState('')
  const [newVisibility, setNewVisibility] = useState<'public' | 'workspace' | 'private'>('workspace')
  const [shareEmail, setShareEmail] = useState('')
  const [shareRole, setShareRole] = useState<'editor' | 'viewer'>('viewer')

  const fetchDocuments = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/documents?user_id=${user?.id}`, {
        headers: { 'X-User-ID': user?.id || '', 'X-Workspace-ID': 'workspace-1' },
      })
      const data = await res.json()
      setDocuments(data.documents || [])
    } catch {
      setError('Failed to fetch documents')
    } finally {
      setLoading(false)
    }
  }, [user?.id])

  useEffect(() => {
    fetchDocuments()
  }, [fetchDocuments])

  const fetchDocument = async (id: string) => {
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/documents/${id}?user_id=${user?.id}`, {
        headers: { 'X-User-ID': user?.id || '', 'X-Workspace-ID': 'workspace-1' },
      })
      const data = await res.json()
      setSelectedDoc(data.document)
      setShares(data.shares || [])
    } catch {
      setError('Failed to fetch document')
    }
  }

  const createDocument = async () => {
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/documents`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': user?.id || '',
          'X-Workspace-ID': 'workspace-1',
        },
        body: JSON.stringify({
          title: newTitle,
          content: newContent,
          visibility: newVisibility,
        }),
      })
      if (!res.ok) {
        const err = await res.json()
        throw new Error(err.message || 'Failed to create')
      }
      setShowCreateModal(false)
      setNewTitle('')
      setNewContent('')
      fetchDocuments()
    } catch (err: unknown) {
      const error = err as Error
      setError(error.message)
    }
  }

  const updateDocument = async () => {
    if (!selectedDoc) return
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/documents/${selectedDoc.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': user?.id || '',
          'X-Workspace-ID': 'workspace-1',
        },
        body: JSON.stringify({
          title: newTitle,
          content: newContent,
          visibility: newVisibility,
        }),
      })
      if (!res.ok) {
        const err = await res.json()
        throw new Error(err.message || 'Failed to update')
      }
      setShowEditModal(false)
      fetchDocuments()
      fetchDocument(selectedDoc.id)
    } catch (err: unknown) {
      const error = err as Error
      setError(error.message)
    }
  }

  const deleteDocument = async (id: string) => {
    if (!confirm('Are you sure you want to delete this document?')) return
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/documents/${id}`, {
        method: 'DELETE',
        headers: { 'X-User-ID': user?.id || '', 'X-Workspace-ID': 'workspace-1' },
      })
      if (!res.ok) {
        const err = await res.json()
        throw new Error(err.message || 'Failed to delete')
      }
      setSelectedDoc(null)
      fetchDocuments()
    } catch (err: unknown) {
      const error = err as Error
      setError(error.message)
    }
  }

  const shareDocument = async () => {
    if (!selectedDoc) return
    try {
      const res = await fetch(`${SAMPLE_API_URL}/api/v1/documents/${selectedDoc.id}/share`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-User-ID': user?.id || '',
          'X-Workspace-ID': 'workspace-1',
        },
        body: JSON.stringify({
          user_id: shareEmail,
          role: shareRole,
        }),
      })
      if (!res.ok) {
        const err = await res.json()
        throw new Error(err.message || 'Failed to share')
      }
      setShowShareModal(false)
      setShareEmail('')
      fetchDocument(selectedDoc.id)
    } catch (err: unknown) {
      const error = err as Error
      setError(error.message)
    }
  }

  const openEditModal = (doc: Document) => {
    setNewTitle(doc.title)
    setNewContent(doc.content)
    setNewVisibility(doc.visibility)
    setShowEditModal(true)
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <div>
          <h1>Documents</h1>
          <p style={{ color: 'var(--text-secondary)' }}>
            ReBAC Demo: Access based on relationships (owner, editor, viewer)
          </p>
        </div>
        <button className="btn btn-primary" style={{ width: 'auto' }} onClick={() => setShowCreateModal(true)}>
          + New Document
        </button>
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
        {/* Document List */}
        <div className="card">
          <div className="card-header">
            <div className="card-title">Your Documents</div>
          </div>

          {loading ? (
            <div className="loading"><div className="spinner" /></div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {documents.map(doc => (
                <div
                  key={doc.id}
                  onClick={() => fetchDocument(doc.id)}
                  style={{
                    padding: '1rem',
                    background: selectedDoc?.id === doc.id ? 'var(--background)' : 'transparent',
                    borderRadius: 'var(--radius)',
                    cursor: 'pointer',
                    border: '1px solid var(--border)',
                  }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                    <div>
                      <div style={{ fontWeight: 500 }}>{doc.title}</div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
                        {doc.visibility} • {doc.status}
                      </div>
                    </div>
                    <div style={{ display: 'flex', gap: '0.25rem' }}>
                      {doc.permissions.can_write && (
                        <span title="Can edit" style={{ color: 'var(--success)' }}>✏️</span>
                      )}
                      {doc.permissions.can_share && (
                        <span title="Can share" style={{ color: 'var(--primary)' }}>🔗</span>
                      )}
                      {doc.permissions.can_delete && (
                        <span title="Can delete" style={{ color: 'var(--danger)' }}>🗑️</span>
                      )}
                    </div>
                  </div>
                </div>
              ))}
              {documents.length === 0 && (
                <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: '2rem' }}>
                  No documents yet
                </p>
              )}
            </div>
          )}
        </div>

        {/* Document Detail */}
        <div className="card">
          <div className="card-header">
            <div className="card-title">
              {selectedDoc ? selectedDoc.title : 'Select a document'}
            </div>
            {selectedDoc && (
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                {selectedDoc.permissions.can_write && (
                  <button className="btn btn-secondary btn-small" onClick={() => openEditModal(selectedDoc)}>
                    Edit
                  </button>
                )}
                {selectedDoc.permissions.can_share && (
                  <button className="btn btn-secondary btn-small" onClick={() => setShowShareModal(true)}>
                    Share
                  </button>
                )}
                {selectedDoc.permissions.can_delete && (
                  <button className="btn btn-danger btn-small" onClick={() => deleteDocument(selectedDoc.id)}>
                    Delete
                  </button>
                )}
              </div>
            )}
          </div>

          {selectedDoc ? (
            <div>
              <div style={{ marginBottom: '1rem' }}>
                <div style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
                  Visibility: <strong>{selectedDoc.visibility}</strong> •
                  Status: <strong>{selectedDoc.status}</strong> •
                  Owner: <strong>{selectedDoc.owner_id}</strong>
                </div>
                <div style={{ padding: '1rem', background: 'var(--background)', borderRadius: 'var(--radius)' }}>
                  {selectedDoc.content || 'No content'}
                </div>
              </div>

              {/* Permissions */}
              <div style={{ marginBottom: '1rem' }}>
                <h4 style={{ marginBottom: '0.5rem' }}>Your Permissions</h4>
                <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                  {Object.entries(selectedDoc.permissions).map(([perm, allowed]) => (
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

              {/* Shares */}
              <div>
                <h4 style={{ marginBottom: '0.5rem' }}>Shared With</h4>
                {shares.length > 0 ? (
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '0.25rem' }}>
                    {shares.map((share, i) => (
                      <div key={i} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem', background: 'var(--background)', borderRadius: '4px' }}>
                        <span>{share.user_id}</span>
                        <span className="member-role">{share.role}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Not shared with anyone</p>
                )}
              </div>
            </div>
          ) : (
            <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: '2rem' }}>
              Select a document to view details
            </p>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <div className="modal-title">Create Document</div>
              <button className="modal-close" onClick={() => setShowCreateModal(false)}>×</button>
            </div>
            <div className="form-group">
              <label>Title</label>
              <input value={newTitle} onChange={e => setNewTitle(e.target.value)} placeholder="Document title" />
            </div>
            <div className="form-group">
              <label>Content</label>
              <textarea
                value={newContent}
                onChange={e => setNewContent(e.target.value)}
                placeholder="Document content..."
                rows={4}
                style={{ width: '100%', padding: '0.75rem', border: '1px solid var(--border)', borderRadius: 'var(--radius)' }}
              />
            </div>
            <div className="form-group">
              <label>Visibility</label>
              <select value={newVisibility} onChange={e => setNewVisibility(e.target.value as typeof newVisibility)} style={{ width: '100%', padding: '0.75rem', border: '1px solid var(--border)', borderRadius: 'var(--radius)' }}>
                <option value="public">Public (anyone in workspace)</option>
                <option value="workspace">Workspace (workspace members)</option>
                <option value="private">Private (only shared users)</option>
              </select>
            </div>
            <div style={{ display: 'flex', gap: '1rem' }}>
              <button className="btn btn-secondary" onClick={() => setShowCreateModal(false)}>Cancel</button>
              <button className="btn btn-primary" onClick={createDocument}>Create</button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Modal */}
      {showEditModal && selectedDoc && (
        <div className="modal-overlay" onClick={() => setShowEditModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <div className="modal-title">Edit Document</div>
              <button className="modal-close" onClick={() => setShowEditModal(false)}>×</button>
            </div>
            <div className="form-group">
              <label>Title</label>
              <input value={newTitle} onChange={e => setNewTitle(e.target.value)} />
            </div>
            <div className="form-group">
              <label>Content</label>
              <textarea
                value={newContent}
                onChange={e => setNewContent(e.target.value)}
                rows={4}
                style={{ width: '100%', padding: '0.75rem', border: '1px solid var(--border)', borderRadius: 'var(--radius)' }}
              />
            </div>
            <div className="form-group">
              <label>Visibility</label>
              <select
                value={newVisibility}
                onChange={e => setNewVisibility(e.target.value as typeof newVisibility)}
                style={{ width: '100%', padding: '0.75rem', border: '1px solid var(--border)', borderRadius: 'var(--radius)' }}
                disabled={selectedDoc.owner_id !== user?.id}
              >
                <option value="public">Public</option>
                <option value="workspace">Workspace</option>
                <option value="private">Private</option>
              </select>
              {selectedDoc.owner_id !== user?.id && (
                <small style={{ color: 'var(--text-secondary)' }}>Only owner can change visibility</small>
              )}
            </div>
            <div style={{ display: 'flex', gap: '1rem' }}>
              <button className="btn btn-secondary" onClick={() => setShowEditModal(false)}>Cancel</button>
              <button className="btn btn-primary" onClick={updateDocument}>Save</button>
            </div>
          </div>
        </div>
      )}

      {/* Share Modal */}
      {showShareModal && selectedDoc && (
        <div className="modal-overlay" onClick={() => setShowShareModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <div className="modal-title">Share Document</div>
              <button className="modal-close" onClick={() => setShowShareModal(false)}>×</button>
            </div>
            <div className="form-group">
              <label>User ID</label>
              <input value={shareEmail} onChange={e => setShareEmail(e.target.value)} placeholder="user-123" />
            </div>
            <div className="form-group">
              <label>Role</label>
              <select value={shareRole} onChange={e => setShareRole(e.target.value as typeof shareRole)} style={{ width: '100%', padding: '0.75rem', border: '1px solid var(--border)', borderRadius: 'var(--radius)' }}>
                <option value="viewer">Viewer (read only)</option>
                <option value="editor">Editor (read & write)</option>
              </select>
            </div>
            <div style={{ display: 'flex', gap: '1rem' }}>
              <button className="btn btn-secondary" onClick={() => setShowShareModal(false)}>Cancel</button>
              <button className="btn btn-primary" onClick={shareDocument}>Share</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
