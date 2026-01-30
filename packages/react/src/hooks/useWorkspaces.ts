import { useSaasAuthContext } from '../contexts/SaasAuthContext'
import type { UseWorkspacesReturn } from '../types'

/**
 * Hook for workspace management
 *
 * @example
 * ```tsx
 * const {
 *   workspaces,
 *   currentWorkspace,
 *   setCurrentWorkspace,
 *   createWorkspace
 * } = useWorkspaces()
 *
 * // Workspace selector
 * <select
 *   value={currentWorkspace?.id}
 *   onChange={(e) => setCurrentWorkspace(e.target.value)}
 * >
 *   {workspaces.map(ws => (
 *     <option key={ws.id} value={ws.id}>{ws.displayName}</option>
 *   ))}
 * </select>
 * ```
 */
export function useWorkspaces(): UseWorkspacesReturn {
  const context = useSaasAuthContext()

  return {
    workspaces: context.workspaces,
    currentWorkspace: context.currentWorkspace,
    isLoading: context.isLoading,
    error: context.error ? { error: context.error.code, message: context.error.message } : null,

    // Actions
    setCurrentWorkspace: context.setCurrentWorkspace,
    createWorkspace: context.createWorkspace,
    refreshWorkspaces: context.refreshWorkspaces,
  }
}
