import { NetworkFunction, HealthStatus, NetworkFunctionType } from '../types/networkFunction'
import { SystemStatus, Metrics, Alarm } from '../types/system'

// API configuration
const API_BASE_URL = process.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'
const KUBERNETES_API_URL = process.env.VITE_K8S_API_URL || 'http://localhost:8080/api/v1/k8s'
const PROMETHEUS_URL = process.env.VITE_PROMETHEUS_URL || 'http://localhost:9090'

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message)
    this.name = 'ApiError'
  }
}

// Generic fetch wrapper with error handling
async function fetchApi<T>(url: string, options?: RequestInit): Promise<T> {
  try {
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      ...options,
    })

    if (!response.ok) {
      throw new ApiError(response.status, `HTTP ${response.status}: ${response.statusText}`)
    }

    return await response.json()
  } catch (error) {
    if (error instanceof ApiError) {
      throw error
    }
    throw new ApiError(0, `Network error: ${error instanceof Error ? error.message : 'Unknown error'}`)
  }
}

export const api = {
  // System Status
  async getSystemStatus(): Promise<SystemStatus> {
    return fetchApi<SystemStatus>(`${API_BASE_URL}/system/status`)
  },

  // Network Function Discovery
  async discoverKubernetesServices(namespaces: string[]): Promise<NetworkFunction[]> {
    const services: NetworkFunction[] = []
    
    for (const namespace of namespaces) {
      try {
        const response = await fetchApi<any>(`${KUBERNETES_API_URL}/namespaces/${namespace}/services`)
        
        for (const service of response.items || []) {
          const nf: NetworkFunction = {
            id: `${namespace}-${service.metadata.name}`,
            name: service.metadata.name,
            namespace,
            type: inferNetworkFunctionType(service.metadata.name, service.metadata.labels),
            status: 'active',
            health: 'healthy',
            version: service.metadata.labels?.version || 'unknown',
            endpoints: extractServiceEndpoints(service),
            labels: service.metadata.labels || {},
            annotations: service.metadata.annotations || {},
            createdAt: new Date(service.metadata.creationTimestamp),
            lastSeen: new Date(),
          }
          services.push(nf)
        }
      } catch (error) {
        console.warn(`Failed to discover services in namespace ${namespace}:`, error)
      }
    }
    
    return services
  },

  async discoverNearRtRicComponents(): Promise<NetworkFunction[]> {
    try {
      const response = await fetchApi<any>(`${API_BASE_URL}/ric/components`)
      
      return response.components?.map((component: any) => ({
        id: `ric-${component.name}`,
        name: component.name,
        namespace: 'oran-nearrt-ric',
        type: component.type as NetworkFunctionType,
        status: component.status,
        health: component.health,
        version: component.version,
        endpoints: component.endpoints || [],
        metrics: component.metrics,
        interfaces: component.interfaces || [],
        createdAt: new Date(component.createdAt),
        lastSeen: new Date(),
      })) || []
    } catch (error) {
      console.warn('Failed to discover Near-RT RIC components:', error)
      return []
    }
  },

  async discoverXApps(): Promise<NetworkFunction[]> {
    try {
      const response = await fetchApi<any>(`${API_BASE_URL}/xapps`)
      
      return response.xapps?.map((xapp: any) => ({
        id: `xapp-${xapp.name}`,
        name: xapp.name,
        namespace: xapp.namespace || 'oran-nearrt-ric',
        type: 'xapp' as NetworkFunctionType,
        status: xapp.status,
        health: xapp.health,
        version: xapp.version,
        endpoints: xapp.endpoints || [],
        config: xapp.config,
        resources: xapp.resources,
        createdAt: new Date(xapp.deployedAt),
        lastSeen: new Date(),
      })) || []
    } catch (error) {
      console.warn('Failed to discover xApps:', error)
      return []
    }
  },

  async discoverSmoComponents(): Promise<NetworkFunction[]> {
    try {
      const response = await fetchApi<any>(`${API_BASE_URL}/smo/components`)
      
      return response.components?.map((component: any) => ({
        id: `smo-${component.name}`,
        name: component.name,
        namespace: 'onap',
        type: component.type as NetworkFunctionType,
        status: component.status,
        health: component.health,
        version: component.version,
        endpoints: component.endpoints || [],
        interfaces: ['a1', 'o1', 'o2'],
        createdAt: new Date(component.deployedAt),
        lastSeen: new Date(),
      })) || []
    } catch (error) {
      console.warn('Failed to discover SMO components:', error)
      return []
    }
  },

  async discoverOranSimulators(): Promise<NetworkFunction[]> {
    try {
      const response = await fetchApi<any>(`${API_BASE_URL}/simulators`)
      
      return response.simulators?.map((simulator: any) => ({
        id: `sim-${simulator.name}`,
        name: simulator.name,
        namespace: simulator.namespace || 'oran-nearrt-ric',
        type: simulator.type as NetworkFunctionType,
        status: simulator.status,
        health: simulator.health,
        version: simulator.version,
        endpoints: simulator.endpoints || [],
        config: simulator.config,
        createdAt: new Date(simulator.createdAt),
        lastSeen: new Date(),
      })) || []
    } catch (error) {
      console.warn('Failed to discover O-RAN simulators:', error)
      return []
    }
  },

  // Health Status
  async getNetworkFunctionHealth(name: string, namespace: string): Promise<{ status: HealthStatus; metrics?: any }> {
    try {
      const response = await fetchApi<any>(`${API_BASE_URL}/health/${namespace}/${name}`)
      return {
        status: response.status,
        metrics: response.metrics,
      }
    } catch (error) {
      return { status: 'unknown' }
    }
  },

  // Network Functions Overview
  async getNetworkFunctionsOverview(): Promise<any> {
    return fetchApi<any>(`${API_BASE_URL}/network-functions/overview`)
  },

  // Real-time Metrics
  async getMetrics(query: string, timeRange?: string): Promise<Metrics> {
    const params = new URLSearchParams({
      query,
      ...(timeRange && { time: timeRange }),
    })
    
    return fetchApi<Metrics>(`${PROMETHEUS_URL}/api/v1/query?${params}`)
  },

  async getMetricsRange(query: string, start: string, end: string, step: string): Promise<Metrics> {
    const params = new URLSearchParams({
      query,
      start,
      end,
      step,
    })
    
    return fetchApi<Metrics>(`${PROMETHEUS_URL}/api/v1/query_range?${params}`)
  },

  // Alarms
  async getAlarms(): Promise<Alarm[]> {
    return fetchApi<Alarm[]>(`${API_BASE_URL}/alarms`)
  },

  async acknowledgeAlarm(alarmId: string): Promise<void> {
    await fetchApi(`${API_BASE_URL}/alarms/${alarmId}/acknowledge`, {
      method: 'POST',
    })
  },

  // Configuration
  async getConfiguration(component: string): Promise<any> {
    return fetchApi<any>(`${API_BASE_URL}/config/${component}`)
  },

  async updateConfiguration(component: string, config: any): Promise<void> {
    await fetchApi(`${API_BASE_URL}/config/${component}`, {
      method: 'PUT',
      body: JSON.stringify(config),
    })
  },

  // xApp Management
  async deployXApp(xappDescriptor: any): Promise<void> {
    await fetchApi(`${API_BASE_URL}/xapps/deploy`, {
      method: 'POST',
      body: JSON.stringify(xappDescriptor),
    })
  },

  async undeployXApp(xappName: string): Promise<void> {
    await fetchApi(`${API_BASE_URL}/xapps/${xappName}/undeploy`, {
      method: 'DELETE',
    })
  },

  // SMO Operations
  async getSmoStatus(): Promise<any> {
    return fetchApi<any>(`${API_BASE_URL}/smo/status`)
  },

  async triggerNetworkSliceCreation(sliceDescriptor: any): Promise<void> {
    await fetchApi(`${API_BASE_URL}/smo/network-slices`, {
      method: 'POST',
      body: JSON.stringify(sliceDescriptor),
    })
  },

  // Interface Testing
  async testE2Interface(nodeId: string): Promise<any> {
    return fetchApi<any>(`${API_BASE_URL}/interfaces/e2/test/${nodeId}`, {
      method: 'POST',
    })
  },

  async testA1Interface(policyId: string): Promise<any> {
    return fetchApi<any>(`${API_BASE_URL}/interfaces/a1/test/${policyId}`, {
      method: 'POST',
    })
  },

  async testO1Interface(managedElementId: string): Promise<any> {
    return fetchApi<any>(`${API_BASE_URL}/interfaces/o1/test/${managedElementId}`, {
      method: 'POST',
    })
  },
}

// Helper functions
function inferNetworkFunctionType(name: string, labels?: Record<string, string>): NetworkFunctionType {
  // Check labels first
  if (labels?.['app.kubernetes.io/component']) {
    const component = labels['app.kubernetes.io/component']
    if (['xapp', 'e2term', 'a1mediator', 'o1mediator', 'appmgr', 'rtmgr', 'submgr'].includes(component)) {
      return component as NetworkFunctionType
    }
  }

  // Infer from name patterns
  if (name.includes('xapp')) return 'xapp'
  if (name.includes('e2term') || name.includes('e2-term')) return 'e2term'
  if (name.includes('a1mediator') || name.includes('a1-mediator')) return 'a1mediator'
  if (name.includes('o1mediator') || name.includes('o1-mediator')) return 'o1mediator'
  if (name.includes('appmgr') || name.includes('app-mgr')) return 'appmgr'
  if (name.includes('rtmgr') || name.includes('routing-manager')) return 'rtmgr'
  if (name.includes('submgr') || name.includes('subscription-manager')) return 'submgr'
  if (name.includes('dashboard')) return 'dashboard'
  if (name.includes('simulator')) return 'simulator'
  if (name.includes('smo')) return 'smo'
  if (name.includes('aai')) return 'aai'
  if (name.includes('so')) return 'so'
  if (name.includes('sdnc')) return 'sdnc'
  if (name.includes('policy')) return 'policy'
  if (name.includes('dmaap')) return 'dmaap'

  return 'unknown'
}

function extractServiceEndpoints(service: any): string[] {
  const endpoints: string[] = []
  
  if (service.spec?.ports) {
    for (const port of service.spec.ports) {
      const protocol = port.protocol?.toLowerCase() || 'tcp'
      const endpoint = `${protocol}://${service.metadata.name}.${service.metadata.namespace}.svc.cluster.local:${port.port}`
      endpoints.push(endpoint)
    }
  }
  
  return endpoints
}