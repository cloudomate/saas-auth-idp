import { Routes, Route, Navigate } from 'react-router-dom'
import { ProtectedRoute } from '@saas-starter/react'

// Auth pages
import { LoginPage } from './pages/auth/LoginPage'
import { RegisterPage } from './pages/auth/RegisterPage'
import { VerifyEmailPage } from './pages/auth/VerifyEmailPage'
import { ForgotPasswordPage } from './pages/auth/ForgotPasswordPage'
import { ResetPasswordPage } from './pages/auth/ResetPasswordPage'
import { OAuthCallbackPage } from './pages/auth/OAuthCallbackPage'

// Onboarding pages
import { SelectPlanPage } from './pages/onboarding/SelectPlanPage'
import { SetupOrgPage } from './pages/onboarding/SetupOrgPage'

// App pages
import { DashboardPage } from './pages/app/DashboardPage'
import { WorkspacesPage } from './pages/app/WorkspacesPage'
import { WorkspaceDetailPage } from './pages/app/WorkspaceDetailPage'
import { SettingsPage } from './pages/app/SettingsPage'
import { DocumentsPage } from './pages/app/DocumentsPage'
import { ProjectsPage } from './pages/app/ProjectsPage'

// Layout
import { AppLayout } from './components/layout/AppLayout'

function App() {
  return (
    <Routes>
      {/* Public auth routes */}
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/verify-email" element={<VerifyEmailPage />} />
      <Route path="/forgot-password" element={<ForgotPasswordPage />} />
      <Route path="/reset-password" element={<ResetPasswordPage />} />
      <Route path="/auth/callback" element={<OAuthCallbackPage />} />

      {/* Onboarding routes (auth required, no tenant) */}
      <Route
        path="/onboarding/plan"
        element={
          <ProtectedRoute fallback={<Navigate to="/login" replace />}>
            <SelectPlanPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/onboarding/setup"
        element={
          <ProtectedRoute fallback={<Navigate to="/login" replace />}>
            <SetupOrgPage />
          </ProtectedRoute>
        }
      />

      {/* Protected app routes */}
      <Route
        path="/"
        element={
          <ProtectedRoute
            fallback={<Navigate to="/login" replace />}
            requireTenant
            tenantSetupFallback={<Navigate to="/onboarding/plan" replace />}
          >
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
