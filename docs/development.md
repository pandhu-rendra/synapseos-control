# Development

## Current phase

Phase 7 MVP quality assurance is complete. The project includes the local
Docker connection, container management, logs, Overview and Host views, runtime
settings, API token protection, and security documentation.

There is no Phase 8 in the current roadmap. New work requires an explicitly
approved feature or post-MVP scope.

## Validation

Backend:

```bash
cd apps/api
gofmt -w .
go test ./...
go vet ./...
```

Frontend:

```bash
cd apps/web
npm install
npm run typecheck
npm run lint
npm run build
```

Automated provider and HTTP tests run without a daemon. Live Docker behavior
should be smoke-tested from a session with access to the user's Docker group.

Set `SYNAPSEOS_API_TOKEN` when starting the API and set the same value as
`VITE_API_TOKEN` for the web app. Health is public for startup checks; all
Docker, container, and settings endpoints require this bearer token. The API
binds to `127.0.0.1:8080` by default.

The API also loads `apps/api/.env` automatically. Copy
`apps/api/.env.example` to `.env` and keep that file untracked. The web app
loads `apps/web/.env` through Vite.
