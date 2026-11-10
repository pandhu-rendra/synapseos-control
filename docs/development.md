# Development

## Current phase

Phase 6 adds the Overview dashboard using real Docker status and info data.
It shows connection state, engine identity, resource counts, stopped-container
attention, and host information. Runtime Docker settings are now available for
the local Unix socket and timeout. Phase 7 covers final quality assurance.

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

Real Docker integration tests cannot run until Docker is available in the
development environment. Isolated provider and HTTP handler tests still run
without a daemon.
