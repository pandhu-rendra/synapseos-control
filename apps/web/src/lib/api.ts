export type ApiEnvelope<T> = {
  data: T
  error: null
}

type ErrorEnvelope<T> = {
  data: T | null
  error: { code: string; message: string } | null
}

export type HealthData = {
  status: string
}

export type DockerStatus = {
  connected: boolean
  api_version?: string
  os_type?: string
}

export type DockerInfo = {
  name: string
  server_version: string
  operating_system: string
  architecture: string
  cpus: number
  memory_bytes: number
  containers: number
  running_containers: number
  stopped_containers: number
  images: number
}

export type DockerSettings = {
  host: string
  timeout_seconds: number
}

export type Container = {
  id: string
  name: string
  image: string
  state: string
  status: string
}

export type ContainerDetails = Container & {
  created: string
  ports: Array<{
    private_port: number
    public_port: number
    type: string
    ip: string
  }>
  networks: string[]
}

export type ContainerAction = 'start' | 'stop' | 'restart'

export type LogLine = {
  stream: 'stdout' | 'stderr'
  text: string
}

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'
const apiToken = import.meta.env.VITE_API_TOKEN ?? ''

function requestHeaders(): HeadersInit {
  return apiToken ? { Authorization: `Bearer ${apiToken}` } : {}
}

export async function fetchHealth(): Promise<HealthData> {
  const response = await fetch(`${apiBaseUrl}/api/v1/health`, { headers: requestHeaders() })
  if (!response.ok) {
    throw new Error(`Health request failed with status ${response.status}`)
  }

  const payload = (await response.json()) as ApiEnvelope<HealthData>
  return payload.data
}

export async function fetchDockerStatus(): Promise<DockerStatus> {
  const response = await fetch(`${apiBaseUrl}/api/v1/docker/status`, { headers: requestHeaders() })
  const payload = await response.json() as ErrorEnvelope<DockerStatus>
  if (!response.ok || payload.error || !payload.data) {
    throw new Error(payload.error?.message ?? `Docker status request failed with status ${response.status}`)
  }

  return payload.data
}

export async function fetchDockerInfo(): Promise<DockerInfo> {
  const response = await fetch(`${apiBaseUrl}/api/v1/docker/info`, { headers: requestHeaders() })
  const payload = await response.json() as ErrorEnvelope<DockerInfo>
  if (!response.ok || payload.error || !payload.data) {
    throw new Error(payload.error?.message ?? `Docker info request failed with status ${response.status}`)
  }

  return payload.data
}

export async function fetchDockerSettings(): Promise<DockerSettings> {
  const response = await fetch(`${apiBaseUrl}/api/v1/settings/docker`, { headers: requestHeaders() })
  const payload = await response.json() as ErrorEnvelope<DockerSettings>
  if (!response.ok || payload.error || !payload.data) {
    throw new Error(payload.error?.message ?? `Docker settings request failed with status ${response.status}`)
  }
  return payload.data
}

export async function saveDockerSettings(settings: DockerSettings): Promise<DockerSettings> {
  const response = await fetch(`${apiBaseUrl}/api/v1/settings/docker`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', ...requestHeaders() },
    body: JSON.stringify(settings),
  })
  const payload = await response.json() as ErrorEnvelope<DockerSettings>
  if (!response.ok || payload.error || !payload.data) {
    throw new Error(payload.error?.message ?? `Docker settings request failed with status ${response.status}`)
  }
  return payload.data
}

export async function fetchContainers(state = ''): Promise<Container[]> {
  const query = state ? `?state=${encodeURIComponent(state)}` : ''
  const response = await fetch(`${apiBaseUrl}/api/v1/containers${query}`, { headers: requestHeaders() })
  const payload = await response.json() as ErrorEnvelope<Container[]>
  if (!response.ok || payload.error || !payload.data) {
    throw new Error(payload.error?.message ?? `Container request failed with status ${response.status}`)
  }

  return payload.data
}

export async function fetchContainerDetails(id: string): Promise<ContainerDetails> {
  const response = await fetch(`${apiBaseUrl}/api/v1/containers/${encodeURIComponent(id)}`, { headers: requestHeaders() })
  const payload = await response.json() as ErrorEnvelope<ContainerDetails>
  if (!response.ok || payload.error || !payload.data) {
    throw new Error(payload.error?.message ?? `Container details request failed with status ${response.status}`)
  }

  return payload.data
}

export async function performContainerAction(id: string, action: ContainerAction): Promise<void> {
  const response = await fetch(
    `${apiBaseUrl}/api/v1/containers/${encodeURIComponent(id)}/${action}`,
    { method: 'POST', headers: requestHeaders() },
  )
  const payload = await response.json() as ErrorEnvelope<{ status: string }>
  if (!response.ok || payload.error) {
    throw new Error(payload.error?.message ?? `Container action failed with status ${response.status}`)
  }
}

export async function fetchContainerLogs(id: string, tail: number): Promise<LogLine[]> {
  const response = await fetch(
    `${apiBaseUrl}/api/v1/containers/${encodeURIComponent(id)}/logs?tail=${tail}`,
    { headers: requestHeaders() },
  )
  const payload = await response.json() as ErrorEnvelope<LogLine[]>
  if (!response.ok || payload.error || !payload.data) {
    throw new Error(payload.error?.message ?? `Container logs request failed with status ${response.status}`)
  }
  return payload.data
}
