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

Docker endpoints will be documented here as they are implemented.
