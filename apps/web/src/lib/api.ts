export type ApiEnvelope<T> = {
  data: T
  error: null
}

export type HealthData = {
  status: string
}

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

export async function fetchHealth(): Promise<HealthData> {
  const response = await fetch(`${apiBaseUrl}/api/v1/health`)
  if (!response.ok) {
    throw new Error(`Health request failed with status ${response.status}`)
  }

  const payload = (await response.json()) as ApiEnvelope<HealthData>
  return payload.data
}
