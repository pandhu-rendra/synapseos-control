# Security

The application manages a privileged Docker connection. Phase 1 connects only
to the configured local Docker endpoint and exposes read-only status and
information endpoints.

Planned safeguards include:

- Keep Docker access in the Go backend.
- Never expose the Docker socket to the browser.
- Do not accept arbitrary Docker API paths or commands.
- Validate identifiers and operation parameters.
- Use request timeouts and context cancellation.
- Do not expose secrets or container environment variables by default.
- Do not recommend world-writable Docker socket permissions.

The API binds to localhost by default and requires `SYNAPSEOS_API_TOKEN` for
all Docker, container, and settings endpoints. Keep the token out of source
control and configure the matching `VITE_API_TOKEN` only in local frontend
environment configuration. Do not expose the API beyond localhost without a
review of authentication, TLS, and network controls.
