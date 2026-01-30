import { useState, useCallback, useEffect, useMemo } from 'react'
import { useSaasAuthContext } from '../contexts/SaasAuthContext'
import type { Container, Membership, HierarchyLevel, UseContainersReturn } from '../types'

/**
 * Hook for managing containers at a specific hierarchy level
 *
 * This is a generic hook that works with any level in your hierarchy.
 * It's the recommended approach for building level-specific UI components.
 *
 * @param levelName - The hierarchy level name (e.g., 'workspace', 'project', 'team')
 *
 * @example
 * ```tsx
 * // Workspace selector (default 2-level hierarchy)
 * const { containers, currentContainer, setCurrentContainer, create } = useContainers('workspace')
 *
 * // Project selector (3-level ML platform hierarchy)
 * const { containers, currentContainer, create } = useContainers('project')
 *
 * // Team management
 * const { containers, create, delete: deleteTeam, listMembers } = useContainers('team')
 * ```
 */
export function useContainers(levelName: string): UseContainersReturn {
  const context = useSaasAuthContext()

  // State
  const [containers, setContainers] = useState<Container[]>([])
  const [currentContainer, setCurrentContainerState] = useState<Container | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [levelConfig, setLevelConfig] = useState<HierarchyLevel | null>(null)

  // Fetch level config from hierarchy
  const fetchLevelConfig = useCallback(async () => {
    try {
      const response = await context.api.get<{
        levels: HierarchyLevel[]
      }>('/api/v1/hierarchy')

      const level = response.levels.find((l) => l.name === levelName)
      if (!level) {
        throw new Error(`Unknown hierarchy level: ${levelName}`)
      }
      setLevelConfig(level)
      return level
    } catch (err) {
      console.error('Failed to fetch level config:', err)
      throw err
    }
  }, [context.api, levelName])

  // Fetch containers
  const fetchContainers = useCallback(
    async (parentId?: string) => {
      setIsLoading(true)
      setError(null)

      try {
        let level = levelConfig
        if (!level) {
          level = await fetchLevelConfig()
        }

        const params = parentId ? `?parent_id=${parentId}` : ''
        const response = await context.api.get<Record<string, Container[]>>(
          `/api/v1/${level.url_path}${params}`
        )

        const containerList = response[level.plural] || []
        setContainers(containerList)

        // Restore current container from storage
        const storedId = context.storage.getItem(`container_${levelName}`)
        const validContainer = containerList.find((c) => c.id === storedId)
        setCurrentContainerState(validContainer || containerList[0] || null)
      } catch (err) {
        setError('Failed to fetch containers')
        console.error('Failed to fetch containers:', err)
      } finally {
        setIsLoading(false)
      }
    },
    [context.api, context.storage, levelConfig, levelName, fetchLevelConfig]
  )

  // Set current container
  const setCurrentContainer = useCallback(
    (container: Container | null) => {
      setCurrentContainerState(container)
      if (container) {
        context.storage.setItem(`container_${levelName}`, container.id)
      } else {
        context.storage.removeItem(`container_${levelName}`)
      }
    },
    [context.storage, levelName]
  )

  // Create container
  const create = useCallback(
    async (name: string, slug?: string, parentId?: string): Promise<Container> => {
      if (!levelConfig) {
        throw new Error('Level configuration not loaded')
      }

      const response = await context.api.post<Record<string, Container>>(
        `/api/v1/${levelConfig.url_path}`,
        {
          name,
          slug,
          parent_id: parentId,
        }
      )

      const newContainer = response[levelConfig.name]
      setContainers((prev) => [...prev, newContainer])
      return newContainer
    },
    [context.api, levelConfig]
  )

  // Delete container
  const deleteContainer = useCallback(
    async (id: string): Promise<void> => {
      if (!levelConfig) {
        throw new Error('Level configuration not loaded')
      }

      await context.api.delete(`/api/v1/${levelConfig.url_path}/${id}`)
      setContainers((prev) => prev.filter((c) => c.id !== id))

      // Clear current if deleted
      if (currentContainer?.id === id) {
        const remaining = containers.filter((c) => c.id !== id)
        setCurrentContainer(remaining[0] || null)
      }
    },
    [context.api, levelConfig, currentContainer, containers, setCurrentContainer]
  )

  // List members
  const listMembers = useCallback(
    async (containerId: string): Promise<Membership[]> => {
      if (!levelConfig) {
        throw new Error('Level configuration not loaded')
      }

      const response = await context.api.get<{ members: Membership[] }>(
        `/api/v1/${levelConfig.url_path}/${containerId}/members`
      )
      return response.members
    },
    [context.api, levelConfig]
  )

  // Add member
  const addMember = useCallback(
    async (containerId: string, email: string, role?: string): Promise<void> => {
      if (!levelConfig) {
        throw new Error('Level configuration not loaded')
      }

      await context.api.post(`/api/v1/${levelConfig.url_path}/${containerId}/members`, {
        email,
        role,
      })
    },
    [context.api, levelConfig]
  )

  // Level info for UI display
  const level = useMemo((): HierarchyLevel => {
    return (
      levelConfig || {
        name: levelName,
        display_name: levelName.charAt(0).toUpperCase() + levelName.slice(1),
        plural: levelName + 's',
        url_path: levelName + 's',
        roles: ['admin', 'member', 'viewer'],
        is_root: false,
      }
    )
  }, [levelConfig, levelName])

  // Initialize on mount
  useEffect(() => {
    if (context.isAuthenticated) {
      fetchContainers()
    }
  }, [context.isAuthenticated]) // eslint-disable-line react-hooks/exhaustive-deps

  return {
    containers,
    currentContainer,
    isLoading,
    error,
    setCurrentContainer,
    create,
    delete: deleteContainer,
    listMembers,
    addMember,
    level,
  }
}
