# Architecture

SynapseOS Control uses three layers:

```text
Browser -> React frontend -> Go API -> Docker provider -> Docker Engine
```

Phase 0 includes the browser, React frontend, and Go API health endpoint.
Docker provider integration begins in Phase 1.

The frontend must never access the Docker socket or Docker Engine API directly.
Future Docker implementations must use a provider abstraction so remote
connections can be added without coupling HTTP handlers to socket handling.
