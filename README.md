# SynapseOS Control

SynapseOS Control is a local-first Docker management application.

Phase 0 provides the Go API foundation and React frontend shell. Docker
integration is intentionally not included yet.

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
