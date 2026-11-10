import { useQuery } from '@tanstack/react-query'
import { NavLink, Route, Routes } from 'react-router-dom'
import { fetchHealth } from './lib/api'

const navigation = [
  { label: 'Overview', path: '/' },
  { label: 'Containers', path: '/containers' },
  { label: 'Host', path: '/host' },
  { label: 'Settings', path: '/settings' },
]

export function App() {
  return (
    <div className="min-h-screen bg-bg text-text">
      <a className="skip-link" href="#main-content">Skip to content</a>
      <header className="border-b border-white/10 bg-surface">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-6 px-4 py-4 sm:px-6">
          <NavLink className="text-lg font-semibold tracking-tight" to="/">
            SynapseOS Control
          </NavLink>
          <nav aria-label="Main navigation" className="flex gap-1 overflow-x-auto">
            {navigation.map((item) => (
              <NavLink
                className={({ isActive }) =>
                  `rounded-md px-3 py-2 text-sm transition-colors ${
                    isActive
                      ? 'bg-accent text-bg'
                      : 'text-muted hover:bg-white/5 hover:text-text'
                  }`
                }
                key={item.path}
                to={item.path}
                end={item.path === '/'}
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6" id="main-content">
        <Routes>
          <Route element={<Overview />} path="/" />
          <Route element={<Placeholder title="Containers" />} path="/containers" />
          <Route element={<Placeholder title="Host" />} path="/host" />
          <Route element={<Placeholder title="Settings" />} path="/settings" />
        </Routes>
      </main>
    </div>
  )
}

function Overview() {
  const health = useQuery({ queryKey: ['health'], queryFn: fetchHealth })

  return (
    <section aria-labelledby="overview-heading">
      <p className="mb-2 text-sm text-accent">Local Docker management</p>
      <h1 className="text-3xl font-semibold tracking-tight" id="overview-heading">
        Overview
      </h1>
      <p className="mt-3 max-w-2xl text-muted">
        The foundation is ready. Docker connection data will appear here in Phase 1.
      </p>
      <div className="mt-8 rounded-lg border border-white/10 bg-surface p-5">
        <h2 className="text-lg font-medium">API connection</h2>
        {health.isPending && <p className="mt-2 text-muted">Checking the API...</p>}
        {health.isError && (
          <p className="mt-2 text-danger">
            Cannot reach the API. Start the Go server and refresh this page.
          </p>
        )}
        {health.data && (
          <p className="mt-2 text-success">
            API is connected. Health status: {health.data.status}.
          </p>
        )}
      </div>
    </section>
  )
}

function Placeholder({ title }: { title: string }) {
  return (
    <section aria-labelledby={`${title.toLowerCase()}-heading`}>
      <h1 className="text-3xl font-semibold tracking-tight" id={`${title.toLowerCase()}-heading`}>
        {title}
      </h1>
      <p className="mt-3 text-muted">This area will be implemented in a later phase.</p>
    </section>
  )
}
