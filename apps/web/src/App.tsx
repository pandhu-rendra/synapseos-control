import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { Link, NavLink, Route, Routes, useParams } from 'react-router-dom'
import {
  fetchContainerDetails,
  fetchContainers,
  fetchDockerInfo,
  fetchDockerSettings,
  fetchDockerStatus,
  fetchContainerLogs,
  performContainerAction,
  saveDockerSettings,
  type Container,
  type ContainerAction,
} from './lib/api'

const navigation = [
  { label: 'Overview', path: '/', mark: '01' },
  { label: 'Containers', path: '/containers', mark: '02' },
  { label: 'Host', path: '/host', mark: '03' },
  { label: 'Settings', path: '/settings', mark: '04' },
]

export function App() {
  return (
    <div className="min-h-screen bg-bg text-text">
      <a className="skip-link" href="#main-content">Skip to content</a>
      <header className="topbar">
        <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between gap-6 px-5 py-4 sm:px-8 lg:px-10">
          <NavLink className="brand" to="/">
            <span aria-hidden="true" className="brand-mark">S</span>
            <span>
              <span className="brand-name">SynapseOS</span>
              <span className="brand-product">Control</span>
            </span>
          </NavLink>
          <div className="topbar-meta">
            <span className="eyebrow">Local host</span>
            <span className="topbar-divider" />
            <span className="mono text-xs text-muted">PHASE 6</span>
          </div>
        </div>
      </header>
      <div className="mx-auto flex w-full max-w-[1440px]">
        <aside className="sidebar">
          <nav aria-label="Main navigation" className="sidebar-nav">
            <span className="eyebrow px-3">Workspace</span>
            {navigation.map((item) => (
              <NavigationItem item={item} key={item.path} />
            ))}
          </nav>
          <div className="sidebar-footer">
            <span className="eyebrow">Connection</span>
            <p className="mt-2 text-sm text-muted">Local Docker provider is active.</p>
          </div>
        </aside>
        <main className="main-content" id="main-content">
          <Routes>
            <Route element={<Overview />} path="/" />
            <Route element={<Containers />} path="/containers" />
            <Route element={<ContainerDetails />} path="/containers/:id" />
            <Route element={<Host />} path="/host" />
            <Route element={<Settings />} path="/settings" />
          </Routes>
        </main>
      </div>
      <nav aria-label="Mobile navigation" className="mobile-nav">
        {navigation.map((item) => (
          <NavigationItem item={item} key={item.path} />
        ))}
      </nav>
    </div>
  )
}

function NavigationItem({
  item,
}: {
  item: (typeof navigation)[number]
}) {
  return (
    <NavLink
      className={({ isActive }) =>
        `nav-item ${isActive ? 'nav-item-active' : ''}`
      }
      end={item.path === '/'}
      to={item.path}
    >
      <span className="nav-item-mark" aria-hidden="true">{item.mark}</span>
      <span>{item.label}</span>
    </NavLink>
  )
}

function Overview() {
  const dockerStatus = useQuery({ queryKey: ['docker-status'], queryFn: fetchDockerStatus })
  const dockerInfo = useQuery({ queryKey: ['docker-info'], queryFn: fetchDockerInfo })
  const isDisconnected = dockerStatus.isError || dockerInfo.isError
  const isChecking = dockerStatus.isPending || dockerInfo.isPending

  return (
    <motion.section
      aria-labelledby="overview-heading"
      className="space-y-8"
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.25, ease: 'easeOut' }}
    >
      <div className="page-heading">
        <div>
          <p className="eyebrow accent-label">Infrastructure overview</p>
          <h1 className="page-title" id="overview-heading">Your control surface.</h1>
          <p className="page-lede">
            A clear view of your local infrastructure, built around the
            Docker Engine on this host.
          </p>
        </div>
        <span className="phase-label">Overview / 06</span>
      </div>
      <div className="status-panel">
        <div className="status-panel-heading">
          <div>
            <p className="eyebrow">System check</p>
            <h2 className="section-title">Docker connection</h2>
          </div>
            <span className={`status-badge ${isDisconnected ? 'status-badge-error' : ''}`}>
            <span aria-hidden="true" className="status-dot" />
              {isChecking ? 'Checking' : isDisconnected ? 'Disconnected' : 'Connected'}
          </span>
        </div>
        <div className="status-panel-body">
            {isChecking && <p className="text-muted">Checking the local Docker daemon...</p>}
            {isDisconnected && (
            <p className="text-danger">
                Cannot reach Docker. Check that the daemon is running and that this app can read the socket.
            </p>
          )}
            {dockerInfo.data && dockerStatus.data && (
            <p className="text-muted">
                {dockerInfo.data.name} is running Docker {dockerInfo.data.server_version} on {dockerInfo.data.operating_system}.
            </p>
          )}
            <span className="mono text-xs text-muted">GET /api/v1/docker/info</span>
        </div>
      </div>
      {dockerInfo.data && (
        <>
          <div className="overview-stats" aria-label="Docker resource counts">
            <MetricCard label="Total containers" value={dockerInfo.data.containers} />
            <MetricCard label="Running" value={dockerInfo.data.running_containers} tone="success" />
            <MetricCard label="Stopped" value={dockerInfo.data.stopped_containers} tone="warning" />
            <MetricCard label="Images" value={dockerInfo.data.images} />
          </div>
          <div className="overview-lower-grid">
            <section className={`attention-panel ${dockerInfo.data.stopped_containers > 0 ? 'attention-panel-warning' : ''}`} aria-labelledby="attention-heading">
              <p className="eyebrow">Needs attention</p>
              <h2 className="section-title" id="attention-heading">
                {dockerInfo.data.stopped_containers > 0 ? 'Stopped containers need review.' : 'No stopped containers reported.'}
              </h2>
              <p className="mt-3 text-sm leading-6 text-muted">
                {dockerInfo.data.stopped_containers > 0
                  ? `${dockerInfo.data.stopped_containers} stopped container${dockerInfo.data.stopped_containers === 1 ? '' : 's'} reported by Docker.`
                  : 'The Docker Engine reports no stopped containers on this host.'}
              </p>
            </section>
            <section className="host-summary" aria-labelledby="overview-links-heading">
              <p className="eyebrow">Next actions</p>
              <h2 className="section-title" id="overview-links-heading">Manage this host</h2>
              <div className="overview-link-list">
                <Link className="overview-link" to="/containers">
                  <span><strong>Containers</strong><small>Inspect and manage workloads</small></span>
                  <span aria-hidden="true">→</span>
                </Link>
                <Link className="overview-link" to="/host">
                  <span><strong>Host</strong><small>Review Docker Engine diagnostics</small></span>
                  <span aria-hidden="true">→</span>
                </Link>
                <Link className="overview-link" to="/settings">
                  <span><strong>Settings</strong><small>Configure the local connection</small></span>
                  <span aria-hidden="true">→</span>
                </Link>
              </div>
            </section>
          </div>
        </>
      )}
    </motion.section>
  )
}

function MetricCard({ label, value, tone = 'default' }: { label: string; value: number | string; tone?: 'default' | 'success' | 'warning' }) {
  return (
    <article className={`metric-card metric-card-${tone}`}>
      <p className="eyebrow">{label}</p>
      <p className="metric-value">{value}</p>
    </article>
  )
}

function Host() {
  const dockerStatus = useQuery({ queryKey: ['docker-status'], queryFn: fetchDockerStatus })
  const dockerInfo = useQuery({ queryKey: ['docker-info'], queryFn: fetchDockerInfo })
  const dockerSettings = useQuery({ queryKey: ['docker-settings'], queryFn: fetchDockerSettings })
  const isDisconnected = dockerStatus.isError || dockerInfo.isError
  const isChecking = dockerStatus.isPending || dockerInfo.isPending

  return (
    <section aria-labelledby="host-heading" className="space-y-8">
      <div className="page-heading">
        <div>
          <p className="eyebrow accent-label">Docker host</p>
          <h1 className="page-title" id="host-heading">Host.</h1>
          <p className="page-lede">
            Inspect the Docker Engine and resources available on this local host.
          </p>
        </div>
        <span className="phase-label">Host / Engine</span>
      </div>
      <div className="status-panel">
        <div className="status-panel-heading">
          <div>
            <p className="eyebrow">Connection status</p>
            <h2 className="section-title">Docker Engine</h2>
          </div>
          <span className={`status-badge ${isDisconnected ? 'status-badge-error' : ''}`}>
            <span aria-hidden="true" className="status-dot" />
            {isChecking ? 'Checking' : isDisconnected ? 'Disconnected' : 'Connected'}
          </span>
        </div>
        <div className="status-panel-body">
          {isChecking && <p className="text-muted">Checking the local Docker daemon...</p>}
          {isDisconnected && (
            <p className="text-danger">
              Cannot reach Docker. Check that the daemon is running and that this app can read the socket.
            </p>
          )}
          {dockerInfo.data && dockerStatus.data && (
            <p className="text-muted">
              Connected to {dockerInfo.data.name} through the local Docker provider.
            </p>
          )}
          <span className="mono text-xs text-muted">GET /api/v1/docker/status</span>
        </div>
      </div>
      {dockerInfo.data && (
        <>
          <div className="host-details-grid host-details-grid-primary">
            <section className="host-summary" aria-labelledby="engine-details-heading">
              <p className="eyebrow">Engine details</p>
              <h2 className="section-title" id="engine-details-heading">Docker runtime</h2>
              <dl className="host-summary-list">
                <div><dt>Hostname</dt><dd className="mono">{dockerInfo.data.name}</dd></div>
                <div><dt>Docker version</dt><dd className="mono">{dockerInfo.data.server_version}</dd></div>
                <div><dt>Operating system</dt><dd>{dockerInfo.data.operating_system}</dd></div>
                <div><dt>Architecture</dt><dd className="mono">{dockerInfo.data.architecture}</dd></div>
                <div><dt>API version</dt><dd className="mono">{dockerStatus.data?.api_version ?? 'Not reported'}</dd></div>
              </dl>
            </section>
            <section className="host-summary" aria-labelledby="connection-details-heading">
              <p className="eyebrow">Connection details</p>
              <h2 className="section-title" id="connection-details-heading">Local provider</h2>
              <dl className="host-summary-list">
                <div><dt>Transport</dt><dd>Unix socket</dd></div>
                <div><dt>Endpoint</dt><dd className="mono break-all">{dockerSettings.data?.host ?? 'Loading...'}</dd></div>
                <div><dt>Request timeout</dt><dd className="mono">{dockerSettings.data ? `${dockerSettings.data.timeout_seconds}s` : 'Loading...'}</dd></div>
                <div><dt>Provider</dt><dd>Local Docker provider</dd></div>
              </dl>
            </section>
          </div>
          <section aria-labelledby="capacity-heading">
            <div className="section-heading-row">
              <div>
                <p className="eyebrow">Engine capacity</p>
                <h2 className="section-title" id="capacity-heading">Available resources</h2>
              </div>
              <span className="mono text-xs text-muted">GET /api/v1/docker/info</span>
            </div>
            <div className="overview-stats">
              <MetricCard label="CPUs" value={dockerInfo.data.cpus} />
              <MetricCard label="Memory" value={formatBytes(dockerInfo.data.memory_bytes)} />
              <MetricCard label="Images" value={dockerInfo.data.images} />
            </div>
          </section>
          <section aria-labelledby="host-containers-heading">
            <div className="section-heading-row">
              <div>
                <p className="eyebrow">Container workload</p>
                <h2 className="section-title" id="host-containers-heading">Container states</h2>
              </div>
              <Link className="back-link" to="/containers">View containers</Link>
            </div>
            <div className="overview-stats">
              <MetricCard label="Total" value={dockerInfo.data.containers} />
              <MetricCard label="Running" value={dockerInfo.data.running_containers} tone="success" />
              <MetricCard label="Stopped" value={dockerInfo.data.stopped_containers} tone="warning" />
            </div>
          </section>
        </>
      )}
    </section>
  )
}

function Settings() {
  const queryClient = useQueryClient()
  const settings = useQuery({ queryKey: ['docker-settings'], queryFn: fetchDockerSettings })
  const [host, setHost] = useState('')
  const [timeout, setTimeoutValue] = useState(10)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const save = useMutation({
    mutationFn: () => saveDockerSettings({ host, timeout_seconds: timeout }),
    onSuccess: async (data) => {
      setHost(data.host)
      setTimeoutValue(data.timeout_seconds)
      setMessage('Docker settings saved. The runtime provider is using the new values.')
      setError('')
      await queryClient.invalidateQueries({ queryKey: ['docker-status'] })
      await queryClient.invalidateQueries({ queryKey: ['docker-info'] })
    },
    onError: (reason: Error) => {
      setError(reason.message)
      setMessage('')
    },
  })

  useEffect(() => {
    if (settings.data) {
      setHost(settings.data.host)
      setTimeoutValue(settings.data.timeout_seconds)
    }
  }, [settings.data])

  return (
    <section aria-labelledby="settings-heading" className="space-y-8">
      <div className="page-heading">
        <div>
          <p className="eyebrow accent-label">Application settings</p>
          <h1 className="page-title" id="settings-heading">Settings.</h1>
          <p className="page-lede">Configure how this runtime reaches the local Docker Engine.</p>
        </div>
        <span className="phase-label">Settings / Runtime</span>
      </div>
      {settings.isPending && <p className="text-muted">Loading current settings...</p>}
      {settings.isError && (
        <div className="empty-panel border-danger/40">
          <span className="phase-label border-danger/50 text-danger">Unavailable</span>
          <h2 className="section-title mt-5">Cannot load settings.</h2>
          <p className="mt-3 text-sm text-muted">Start the API server, then refresh this page.</p>
        </div>
      )}
      {settings.data && (
        <form className="settings-panel" onSubmit={(event) => { event.preventDefault(); save.mutate() }}>
          <div className="settings-panel-heading">
            <div>
              <p className="eyebrow">Docker connection</p>
              <h2 className="section-title">Local engine</h2>
            </div>
            <span className="phase-label">Local only</span>
          </div>
          <div className="settings-fields">
            <label className="settings-field">
              <span className="eyebrow">Unix socket endpoint</span>
              <input
                className="settings-input mono"
                onChange={(event) => setHost(event.target.value)}
                pattern="unix:///.+"
                required
                value={host}
              />
              <span className="field-help">Remote TCP and SSH endpoints are disabled in the MVP.</span>
            </label>
            <label className="settings-field">
              <span className="eyebrow">Request timeout (seconds)</span>
              <input
                className="settings-input mono"
                max={120}
                min={1}
                onChange={(event) => setTimeoutValue(Number(event.target.value))}
                required
                type="number"
                value={timeout}
              />
              <span className="field-help">Allowed range: 1 to 120 seconds.</span>
            </label>
          </div>
          <div className="settings-actions">
            <button className="action-button action-button-primary" disabled={save.isPending} type="submit">
              {save.isPending ? 'Saving...' : 'Save and test connection'}
            </button>
            {message && <p aria-live="polite" className="text-sm text-success">{message}</p>}
            {error && <p aria-live="polite" className="text-sm text-danger">{error}</p>}
          </div>
        </form>
      )}
    </section>
  )
}

function formatBytes(bytes: number) {
  if (bytes < 1024 ** 3) return `${Math.round(bytes / 1024 ** 2)} MB`
  return `${(bytes / 1024 ** 3).toFixed(1)} GB`
}

function Containers() {
  const [state, setState] = useState('')
  const containers = useQuery({
    queryKey: ['containers', state],
    queryFn: () => fetchContainers(state),
  })

  return (
    <section aria-labelledby="containers-heading" className="space-y-8">
      <div className="page-heading">
        <div>
          <p className="eyebrow accent-label">Docker resources</p>
          <h1 className="page-title" id="containers-heading">Containers.</h1>
          <p className="page-lede">
            A live list of containers reported by your local Docker Engine.
          </p>
        </div>
        <span className="phase-label">Phase / 02</span>
      </div>
      <div className="flex flex-col gap-3 border-b border-white/10 pb-4 sm:flex-row sm:items-center sm:justify-between">
        <p className="eyebrow">Filter by state</p>
        <label className="sr-only" htmlFor="container-state">Container state</label>
        <select
          className="state-select"
          id="container-state"
          onChange={(event) => setState(event.target.value)}
          value={state}
        >
          <option value="">All containers</option>
          <option value="running">Running</option>
          <option value="exited">Exited</option>
          <option value="paused">Paused</option>
          <option value="created">Created</option>
          <option value="restarting">Restarting</option>
          <option value="dead">Dead</option>
        </select>
      </div>
      {containers.isPending && <p className="text-muted">Loading containers...</p>}
      {containers.isError && (
        <div className="empty-panel border-danger/40">
          <span className="phase-label border-danger/50 text-danger">Disconnected</span>
          <h2 className="section-title mt-5">Cannot load containers.</h2>
          <p className="mt-3 max-w-lg text-sm leading-6 text-muted">
            Check that Docker is running and that this app can read the Docker socket.
          </p>
        </div>
      )}
      {containers.data && containers.data.length === 0 && (
        <div className="empty-panel">
          <span className="phase-label">Empty host</span>
          <h2 className="section-title mt-5">No containers on this host.</h2>
          <p className="mt-3 max-w-lg text-sm leading-6 text-muted">
            Start a container with Docker, then refresh this page.
          </p>
        </div>
      )}
      {containers.data && containers.data.length > 0 && (
        <div className="container-list" aria-live="polite">
          {containers.data.map((container) => (
            <ContainerCard container={container} key={container.id} />
          ))}
        </div>
      )}
    </section>
  )
}

function ContainerCard({ container }: { container: Container }) {
  const isRunning = container.state === 'running'
  return (
    <article className={`container-card ${isRunning ? 'container-card-running' : ''}`}>
      <div className="container-card-main">
        <div className="container-state">
          <span aria-hidden="true" className="status-dot" />
          <span>{container.state}</span>
        </div>
        <Link className="container-name container-card-link" to={`/containers/${container.id}`}>
          {container.name || 'Unnamed container'}
        </Link>
        <p className="container-image mono">{container.image}</p>
      </div>
      <div className="container-card-meta">
        <span className="mono container-id">{container.id.slice(0, 12)}</span>
        <span className="container-status">{container.status}</span>
      </div>
    </article>
  )
}

function ContainerDetails() {
  const { id = '' } = useParams()
  const details = useQuery({
    queryKey: ['container', id],
    queryFn: () => fetchContainerDetails(id),
    enabled: Boolean(id),
  })

  return (
    <section aria-labelledby="container-details-heading" className="space-y-8">
      <Link className="back-link" to="/containers">← Back to containers</Link>
      {details.isPending && <p className="text-muted">Loading container details...</p>}
      {details.isError && (
        <div className="empty-panel border-danger/40">
          <span className="phase-label border-danger/50 text-danger">Unavailable</span>
          <h1 className="section-title mt-5" id="container-details-heading">Cannot load this container.</h1>
          <p className="mt-3 max-w-lg text-sm leading-6 text-muted">
            Check the container identifier and confirm that Docker is available.
          </p>
        </div>
      )}
      {details.data && (
        <>
          <div className="page-heading">
            <div>
              <p className="eyebrow accent-label">Container details</p>
              <h1 className="page-title" id="container-details-heading">
                {details.data.name || 'Unnamed container'}
              </h1>
              <p className="page-lede mono">{details.data.id}</p>
            </div>
            <span className="phase-label">{details.data.state}</span>
          </div>
          <ContainerActions
            container={details.data}
            onCompleted={() => {
              void details.refetch()
            }}
          />
          <div className="details-grid">
            <DetailItem label="Image" value={details.data.image} mono />
            <DetailItem label="Created" value={details.data.created} mono />
            <DetailItem label="Status" value={details.data.status} />
            <DetailItem label="Networks" value={details.data.networks.join(', ') || 'None'} />
          </div>
          <section aria-labelledby="ports-heading" className="details-section">
            <p className="eyebrow">Networking</p>
            <h2 className="section-title" id="ports-heading">Port mappings</h2>
            {details.data.ports.length === 0 ? (
              <p className="mt-4 text-sm text-muted">No published ports reported.</p>
            ) : (
              <ul className="port-list">
                {details.data.ports.map((port, index) => (
                  <li className="port-item mono" key={`${port.privatePort}-${port.publicPort}-${index}`}>
                    {port.ip || '*'}:{port.publicPort || 'internal'} → {port.privatePort}/{port.type}
                  </li>
                ))}
              </ul>
            )}
          </section>
          <ContainerLogs containerId={details.data.id} />
        </>
      )}
    </section>
  )
}

function ContainerLogs({ containerId }: { containerId: string }) {
  const [tail, setTail] = useState(200)
  const [followTail, setFollowTail] = useState(true)
  const logs = useQuery({
    queryKey: ['container-logs', containerId, tail],
    queryFn: () => fetchContainerLogs(containerId, tail),
    refetchInterval: followTail ? 5000 : false,
  })

  return (
    <section aria-labelledby="logs-heading" className="details-section">
      <div className="logs-header">
        <div>
          <p className="eyebrow">Output</p>
          <h2 className="section-title" id="logs-heading">Container logs</h2>
        </div>
        <div className="logs-controls">
          <label className="logs-control">
            <span className="eyebrow">Tail</span>
            <select
              aria-label="Log tail length"
              className="state-select"
              onChange={(event) => setTail(Number(event.target.value))}
              value={tail}
            >
              <option value="50">50 lines</option>
              <option value="200">200 lines</option>
              <option value="500">500 lines</option>
              <option value="1000">1000 lines</option>
            </select>
          </label>
          <label className="logs-toggle">
            <input
              checked={followTail}
              onChange={(event) => setFollowTail(event.target.checked)}
              type="checkbox"
            />
            Follow tail
          </label>
        </div>
      </div>
      {logs.isPending && <p className="mt-4 text-sm text-muted">Loading logs...</p>}
      {logs.isError && (
        <p aria-live="polite" className="mt-4 text-sm text-danger">
          Cannot load logs. Check that Docker is available and the container can be inspected.
        </p>
      )}
      {logs.data && logs.data.length === 0 && (
        <p className="mt-4 text-sm text-muted">No log output yet.</p>
      )}
      {logs.data && logs.data.length > 0 && (
        <div aria-label="Container log output" className="log-viewer" role="region" tabIndex={0}>
          {logs.data.map((line, index) => (
            <div className="log-line" key={`${line.stream}-${index}`}>
              <span className={`log-stream log-stream-${line.stream}`}>{line.stream}</span>
              <span className="log-text">{line.text}</span>
            </div>
          ))}
        </div>
      )}
    </section>
  )
}

function ContainerActions({
  container,
  onCompleted,
}: {
  container: Container
  onCompleted: () => void
}) {
  const queryClient = useQueryClient()
  const [actionError, setActionError] = useState('')
  const mutation = useMutation({
    mutationFn: (action: ContainerAction) => performContainerAction(container.id, action),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['containers'] })
      onCompleted()
    },
    onError: (error: Error) => setActionError(error.message),
  })

  const runAction = (action: ContainerAction) => {
    if ((action === 'stop' || action === 'restart') &&
      !window.confirm(`${action === 'stop' ? 'Stop' : 'Restart'} ${container.name || 'this container'}?`)) {
      return
    }
    setActionError('')
    mutation.mutate(action)
  }

  const pendingAction = mutation.variables
  const isRunning = container.state === 'running'
  return (
    <section aria-labelledby="actions-heading" className="action-panel">
      <div>
        <p className="eyebrow">Lifecycle</p>
        <h2 className="section-title" id="actions-heading">Container actions</h2>
      </div>
      <div className="action-buttons">
        <button
          className="action-button action-button-primary"
          disabled={mutation.isPending || isRunning}
          onClick={() => runAction('start')}
          title={isRunning ? 'Already running' : 'Start container'}
          type="button"
        >
          {pendingAction === 'start' ? 'Starting...' : isRunning ? 'Already running' : 'Start'}
        </button>
        <button
          className="action-button"
          disabled={mutation.isPending || !isRunning}
          onClick={() => runAction('stop')}
          title={!isRunning ? 'Already stopped' : 'Stop container'}
          type="button"
        >
          {pendingAction === 'stop' ? 'Stopping...' : !isRunning ? 'Already stopped' : 'Stop'}
        </button>
        <button
          className="action-button"
          disabled={mutation.isPending}
          onClick={() => runAction('restart')}
          type="button"
        >
          {pendingAction === 'restart' ? 'Restarting...' : 'Restart'}
        </button>
      </div>
      {actionError && <p aria-live="polite" className="text-sm text-danger">{actionError}</p>}
      {mutation.isSuccess && !mutation.isPending && (
        <p aria-live="polite" className="text-sm text-success">Container state updated. Docker confirmed the action.</p>
      )}
    </section>
  )
}

function DetailItem({ label, value, mono = false }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="detail-item">
      <p className="eyebrow">{label}</p>
      <p className={`mt-3 break-words text-sm text-text ${mono ? 'mono' : ''}`}>{value || 'Not reported'}</p>
    </div>
  )
}
