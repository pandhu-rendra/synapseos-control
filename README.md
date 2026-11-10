# SynapseOS Control

SynapseOS Control is a local-first Docker management application.

The MVP is complete through Phase 7. It provides a Go API foundation, React
frontend shell, a real local Docker Engine connection, live container listing,
non-secret container details, lifecycle actions, bounded container logs,
real-data Overview and Host views, runtime settings, and API token protection.

## Requirements

- Go 1.24 or newer
- Node.js 20 or newer
- npm

Docker is required starting with Phase 1.

## Run locally

Setup the Token:
```bash
TOKEN=$(openssl rand -hex 32)
echo "$TOKEN"
```

Start the API:

```bash
cd apps/api
export SYNAPSEOS_API_TOKEN='your-token'
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

Set `SYNAPSEOS_API_TOKEN` for the API and the matching `VITE_API_TOKEN` in
the frontend environment. The API binds to localhost by default.

Set `VITE_API_BASE_URL` when the API uses another address. See
`apps/web/.env.example`.

## Project status

The project is being implemented incrementally. See
`docs/development.md` for the current phase and validation commands.
