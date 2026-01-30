import { Routes, Route, Navigate } from 'react-router-dom'
import { useAdminAuth } from './contexts/AdminAuthContext'
import { AdminLayout } from './components/AdminLayout'
import { LoginPage } from './pages/LoginPage'
import { ChangePasswordPage } from './pages/ChangePasswordPage'
import { DashboardPage } from './pages/DashboardPage'
import { UsersPage } from './pages/UsersPage'
import { TenantsPage } from './pages/TenantsPage'

// Protected Route wrapper - requires admin privileges
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading, isAdmin, requiresPasswordChange } = useAdminAuth()

  if (isLoading) {
    return (
      <div className="login-container">
        <div className="login-card">
          <div className="loading">Loading...</div>
        </div>
      </div>
    )
  }

  if (!isAuthenticated || !isAdmin) {
    return <Navigate to="/login" replace />
  }

  // Redirect to password change if required
  if (requiresPasswordChange) {
    return <Navigate to="/change-password" replace />
  }

  return <>{children}</>
}

// Route that requires auth but allows password change
function AuthenticatedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading, isAdmin } = useAdminAuth()

  if (isLoading) {
    return (
      <div className="login-container">
        <div className="login-card">
          <div className="loading">Loading...</div>
        </div>
      </div>
    )
  }

  if (!isAuthenticated || !isAdmin) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function App() {
  return (
    <Routes>
      {/* Public routes */}
      <Route path="/login" element={<LoginPage />} />

      {/* Password change route - requires auth but not password change completion */}
      <Route
        path="/change-password"
        element={
          <AuthenticatedRoute>
            <ChangePasswordPage />
          </AuthenticatedRoute>
        }
      />

      {/* Protected admin routes - requires password change to be completed */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AdminLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<DashboardPage />} />
        <Route path="users" element={<UsersPage />} />
        <Route path="tenants" element={<TenantsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
