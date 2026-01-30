import { useState, useCallback, useEffect, useMemo } from 'react'
import { useSaasAuthContext } from '../contexts/SaasAuthContext'
import type {
  HierarchyConfig,
  HierarchyLevel,
  Container,
  Membership,
  UseHierarchyReturn,
} from '../types'

/**
 * Hook for hierarchy configuration and container operations
 *
 * This hook provides access to the pluggable hierarchy system, allowing
 * applications to work with any hierarchy depth (tenant → workspace,
 * org → team → project, etc.)
 *
 * @example
 * ```tsx
 * const {
 *   config,
 *   rootLevel,
 *   leafLevel,
 *   currentRoot,
 *   currentLeaf,
 *   listContainers,
 *   createContainer,
 *   setCurrentContainer
 * } = useHierarchy()
 *
 * // Display hierarchy levels
 * config?.levels.map(level => (
 *   <div key={level.name}>{level.display_name}</div>
 * ))
 *
 * // Create a container at any level
 * await createContainer('project', 'My Project', 'my-project', parentId)
 * ```
 */
export function useHierarchy(): UseHierarchyReturn {
  const context = useSaasAuthContext()

  // Hierarchy state
  const [config, setConfig] = useState<HierarchyConfig | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [currentContainers, setCurrentContainers] = useState<Record<string, Container | null>>({})

  // Derived values
  const rootLevel = useMemo(() => {
    if (!config) return null
    return config.levels.find((l) => l.is_root) || null
  }, [config])

  const leafLevel = useMemo(() => {
    if (!config) return null
    return config.levels.find((l) => l.name === config.leaf_level) || null
  }, [config])

  const currentRoot = useMemo(() => {
    if (!rootLevel) return null
    return currentContainers[rootLevel.name] || null
  }, [rootLevel, currentContainers])

  const currentLeaf = useMemo(() => {
    if (!leafLevel) return null
    return currentContainers[leafLevel.name] || null
  }, [leafLevel, currentContainers])

  // Get a specific level config
  const getLevel = useCallback(
    (name: string): HierarchyLevel | undefined => {
      return config?.levels.find((l) => l.name === name)
    },
    [config]
  )

  // Fetch hierarchy configuration
  const fetchConfig = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await context.api.get<HierarchyConfig>('/api/v1/hierarchy')
      setConfig(response)
    } catch (err) {
      setError('Failed to fetch hierarchy configuration')
      console.error('Failed to fetch hierarchy config:', err)
    } finally {
      setIsLoading(false)
    }
  }, [context.api])

  // List containers at a level
  const listContainers = useCallback(
    async (level: string, parentId?: string): Promise<Container[]> => {
      const levelConfig = getLevel(level)
      if (!levelConfig) {
        throw new Error(`Unknown hierarchy level: ${level}`)
      }

      const params = parentId ? `?parent_id=${parentId}` : ''
      const response = await context.api.get<Record<string, Container[]>>(
        `/api/v1/${levelConfig.url_path}${params}`
      )

      // Response uses the plural name as key (e.g., { workspaces: [...] })
      return response[levelConfig.plural] || []
    },
    [context.api, getLevel]
  )

  // Create a container
  const createContainer = useCallback(
    async (level: string, name: string, slug?: string, parentId?: string): Promise<Container> => {
      const levelConfig = getLevel(level)
      if (!levelConfig) {
        throw new Error(`Unknown hierarchy level: ${level}`)
      }

      const response = await context.api.post<Record<string, Container>>(
        `/api/v1/${levelConfig.url_path}`,
        {
          name,
          slug,
          parent_id: parentId,
        }
      )

      // Response uses the singular name as key (e.g., { workspace: {...} })
      return response[levelConfig.name]
    },
    [context.api, getLevel]
  )

  // Get a specific container
  const getContainer = useCallback(
    async (level: string, idOrSlug: string): Promise<Container> => {
      const levelConfig = getLevel(level)
      if (!levelConfig) {
        throw new Error(`Unknown hierarchy level: ${level}`)
      }

      const response = await context.api.get<Container>(
        `/api/v1/${levelConfig.url_path}/${idOrSlug}`
      )
      return response
    },
    [context.api, getLevel]
  )

  // Delete a container
  const deleteContainer = useCallback(
    async (level: string, id: string): Promise<void> => {
      const levelConfig = getLevel(level)
      if (!levelConfig) {
        throw new Error(`Unknown hierarchy level: ${level}`)
      }

      await context.api.delete(`/api/v1/${levelConfig.url_path}/${id}`)
    },
    [context.api, getLevel]
  )

  // List members of a container
  const listMembers = useCallback(
    async (level: string, containerId: string): Promise<Membership[]> => {
      const levelConfig = getLevel(level)
      if (!levelConfig) {
        throw new Error(`Unknown hierarchy level: ${level}`)
      }

      const response = await context.api.get<{ members: Membership[] }>(
        `/api/v1/${levelConfig.url_path}/${containerId}/members`
      )
      return response.members
    },
    [context.api, getLevel]
  )

  // Add a member to a container
  const addMember = useCallback(
    async (level: string, containerId: string, email: string, role?: string): Promise<void> => {
      const levelConfig = getLevel(level)
      if (!levelConfig) {
        throw new Error(`Unknown hierarchy level: ${level}`)
      }

      await context.api.post(`/api/v1/${levelConfig.url_path}/${containerId}/members`, {
        email,
        role,
      })
    },
    [context.api, getLevel]
  )

  // Set current container for a level
  const setCurrentContainer = useCallback(
    (level: string, container: Container | null) => {
      setCurrentContainers((prev) => ({
        ...prev,
        [level]: container,
      }))

      // Store in localStorage for persistence
      if (container) {
        context.storage.setItem(`container_${level}`, container.id)
      } else {
        context.storage.removeItem(`container_${level}`)
      }
    },
    [context.storage]
  )

  // Initialize - fetch config on mount
  useEffect(() => {
    if (context.isAuthenticated) {
      fetchConfig()
    }
  }, [context.isAuthenticated]) // eslint-disable-line react-hooks/exhaustive-deps

  return {
    // State
    config,
    isLoading,
    error,
    currentContainers,

    // Actions
    fetchConfig,
    listContainers,
    createContainer,
    getContainer,
    deleteContainer,
    listMembers,
    addMember,
    setCurrentContainer,

    // Convenience getters
    rootLevel,
    leafLevel,
    getLevel,
    currentRoot,
    currentLeaf,
  }
}
