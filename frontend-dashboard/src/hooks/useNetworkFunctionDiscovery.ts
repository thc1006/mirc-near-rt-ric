import { useQuery } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { api } from '../services/api'
import { NetworkFunction, DiscoveryStatus } from '../types/networkFunction'

export interface NetworkFunctionDiscoveryResult {
  networkFunctions: NetworkFunction[]
  isLoading: boolean
  error: Error | null
  status: DiscoveryStatus
  lastDiscovery: Date | null
  refresh: () => void
}

export function useNetworkFunctionDiscovery(): NetworkFunctionDiscoveryResult {
  const [lastDiscovery, setLastDiscovery] = useState<Date | null>(null)
  const [status, setStatus] = useState<DiscoveryStatus>('idle')

  // Auto-discovery query that runs every 30 seconds
  const {
    data: networkFunctions = [],
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['network-functions', 'discovery'],
    queryFn: async () => {
      setStatus('discovering')
      try {
        // Discover Kubernetes services in O-RAN namespaces
        const kubernetesServices = await api.discoverKubernetesServices([
          'oran-nearrt-ric',
          'onap',
          'xapp-dashboard',
          'smo-onap',
        ])

        // Discover Near-RT RIC components via E2 interface
        const ricComponents = await api.discoverNearRtRicComponents()

        // Discover xApps through App Manager
        const xapps = await api.discoverXApps()

        // Discover SMO components via ONAP APIs
        const smoComponents = await api.discoverSmoComponents()

        // Discover O-RAN simulators
        const simulators = await api.discoverOranSimulators()

        // Combine all discovered network functions
        const allNetworkFunctions: NetworkFunction[] = [
          ...kubernetesServices,
          ...ricComponents,
          ...xapps,
          ...smoComponents,
          ...simulators,
        ]

        // Deduplicate and enrich with health status
        const uniqueNetworkFunctions = await enrichWithHealthStatus(
          deduplicateNetworkFunctions(allNetworkFunctions)
        )

        setLastDiscovery(new Date())
        setStatus('completed')
        
        return uniqueNetworkFunctions
      } catch (err) {
        setStatus('error')
        throw err
      }
    },
    refetchInterval: 30000, // Refresh every 30 seconds
    staleTime: 15000, // Consider data stale after 15 seconds
    retry: 3,
    retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
  })

  // Manual refresh function
  const refresh = () => {
    setStatus('discovering')
    refetch()
  }

  // Update discovery status when loading state changes
  useEffect(() => {
    if (isLoading && status === 'idle') {
      setStatus('discovering')
    } else if (!isLoading && !error && status === 'discovering') {
      setStatus('completed')
    } else if (error && status === 'discovering') {
      setStatus('error')
    }
  }, [isLoading, error, status])

  return {
    networkFunctions,
    isLoading,
    error: error as Error | null,
    status,
    lastDiscovery,
    refresh,
  }
}

// Helper function to deduplicate network functions based on name and namespace
function deduplicateNetworkFunctions(networkFunctions: NetworkFunction[]): NetworkFunction[] {
  const seen = new Set<string>()
  return networkFunctions.filter((nf) => {
    const key = `${nf.namespace}-${nf.name}-${nf.type}`
    if (seen.has(key)) {
      return false
    }
    seen.add(key)
    return true
  })
}

// Helper function to enrich network functions with real-time health status
async function enrichWithHealthStatus(networkFunctions: NetworkFunction[]): Promise<NetworkFunction[]> {
  const healthChecks = networkFunctions.map(async (nf) => {
    try {
      const health = await api.getNetworkFunctionHealth(nf.name, nf.namespace)
      return {
        ...nf,
        health: health.status,
        lastHealthCheck: new Date(),
        metrics: health.metrics,
      }
    } catch (error) {
      console.warn(`Failed to get health status for ${nf.name}:`, error)
      return {
        ...nf,
        health: 'unknown' as const,
        lastHealthCheck: new Date(),
      }
    }
  })

  return Promise.all(healthChecks)
}