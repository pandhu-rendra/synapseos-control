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
