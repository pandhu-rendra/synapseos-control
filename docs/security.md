# Security

The application will manage a privileged Docker connection. Phase 0 does not
connect to Docker or expose state-changing operations.

Planned safeguards include:

- Keep Docker access in the Go backend.
- Never expose the Docker socket to the browser.
- Do not accept arbitrary Docker API paths or commands.
- Validate identifiers and operation parameters.
- Use request timeouts and context cancellation.
- Do not expose secrets or container environment variables by default.
- Do not recommend world-writable Docker socket permissions.
