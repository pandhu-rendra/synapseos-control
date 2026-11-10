# SynapseOS Control

SynapseOS Control is a local-first Docker management application.

Phase 6 provides a Go API foundation, React frontend shell, a real local
Docker Engine connection, live container listing, non-secret container
details, lifecycle actions, bounded container logs, and a real-data overview
dashboard. Runtime settings can update the local Unix socket and request
timeout in memory. Phase 7 covers final quality assurance.

## Requirements

- Go 1.24 or newer
- Node.js 20 or newer
- npm

Docker is required starting with Phase 1.

## Run locally

Start the API:

```bash
cd apps/api
go run ./cmd/server
```

Start the web app in another terminal:

```bash
cd apps/web
npm install
npm run dev
```

The API listens on `http://localhost:8080` and the web app on
`http://localhost:5173`.

Set `VITE_API_BASE_URL` when the API uses another address. See
`apps/web/.env.example`.

## Project status

The project is being implemented incrementally. See
`docs/development.md` for the current phase and validation commands.
