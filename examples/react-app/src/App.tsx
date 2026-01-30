import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './contexts/AuthContext'

// Auth pages
import { LoginPage } from './pages/auth/LoginPage'
import { RegisterPage } from './pages/auth/RegisterPage'
import { OAuthCallbackPage } from './pages/auth/OAuthCallbackPage'

// App pages
import { DashboardPage } from './pages/app/DashboardPage'
import { WorkspacesPage } from './pages/app/WorkspacesPage'
import { WorkspaceDetailPage } from './pages/app/WorkspaceDetailPage'
import { SettingsPage } from './pages/app/SettingsPage'
import { DocumentsPage } from './pages/app/DocumentsPage'
import { ProjectsPage } from './pages/app/ProjectsPage'

// Layout
import { AppLayout } from './components/layout/AppLayout'

// Protected Route wrapper
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="loading-container">
        <div className="loading">Loading...</div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function App() {
  return (
    <Routes>
      {/* Public auth routes */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/callback" element={<OAuthCallbackPage />} />
      <Route path="/auth/callback" element={<OAuthCallbackPage />} />

      {/* Protected app routes */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="workspaces" element={<WorkspacesPage />} />
        <Route path="workspaces/:id" element={<WorkspaceDetailPage />} />
        <Route path="documents" element={<DocumentsPage />} />
        <Route path="projects" element={<ProjectsPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>

      {/* Catch all */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
