# API

The API is versioned under `/api/v1` and uses a consistent response envelope.

## Health

`GET /api/v1/health`

Successful response:

```json
{
  "data": { "status": "ok" },
  "error": null
}
```

## Docker status

`GET /api/v1/docker/status`

Returns the result of a real Docker Engine ping.

```json
{
  "data": {
    "connected": true,
    "api_version": "1.45",
    "os_type": "linux"
  },
  "error": null
}
```

When Docker is unavailable, the endpoint returns `503` with
`DOCKER_CONNECTION_FAILED`. Socket access failures return
`DOCKER_PERMISSION_DENIED`.

## Docker information

`GET /api/v1/docker/info`

Returns selected, non-secret host information from Docker Engine, including
engine version, operating system, architecture, CPU count, memory, container
counts, and image count.

## Docker settings

`GET /api/v1/settings/docker` returns the current runtime Docker connection
settings. `PUT /api/v1/settings/docker` replaces them and creates a new local
provider. Settings are held in memory and return to environment defaults when
the API restarts.

Only Unix socket endpoints are accepted. Remote TCP and SSH connections are
not part of the MVP.

```json
{
  "data": {
    "host": "unix:///var/run/docker.sock",
    "timeout_seconds": 10
  },
  "error": null
}
```

## Containers

`GET /api/v1/containers`

Returns all containers from Docker Engine. Use the optional `state` query
parameter to filter by `created`, `running`, `paused`, `restarting`, `removing`,
`exited`, or `dead`.

```json
{
  "data": [
    {
      "id": "abc123",
      "name": "web",
      "image": "nginx:latest",
      "state": "running",
      "status": "Up 2 minutes"
    }
  ],
  "error": null
}
```

## Container details

`GET /api/v1/containers/:id`

Returns non-secret inspection data for one container, including image, state,
creation time, port mappings, and network names. Environment variables and
other sensitive configuration are intentionally excluded.

Invalid identifiers return `400`. A missing container returns `404`.

## Container actions

The following endpoints perform the corresponding Docker Engine operation:

- `POST /api/v1/containers/:id/start`
- `POST /api/v1/containers/:id/stop`
- `POST /api/v1/containers/:id/restart`

Actions return `200` only after Docker confirms the operation. Unsupported
actions return `404`, invalid identifiers return `400`, and missing containers
return `404`.

## Container logs

`GET /api/v1/containers/:id/logs?tail=200`

Returns bounded stdout and stderr lines. `tail` must be between `1` and `1000`
and defaults to `200`. The endpoint handles both running and stopped
containers.
