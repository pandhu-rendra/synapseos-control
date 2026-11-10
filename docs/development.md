# Development

## Current phase

Phase 0 establishes the Go API, React/Vite frontend, routing, configuration,
logging, health communication, and documentation. Docker integration is not
implemented yet.

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

Docker integration tests cannot run until Docker is available in the
development environment.
