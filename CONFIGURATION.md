# Configuration

Environment variables configure the API service.

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `MYSQL_DSN` | yes | — | Marketplace MySQL DSN |
| `API_PORT` | no | `8092` | HTTP listen port |
| `OCTO_API_URL` | when auth enabled | empty | `octo-server` API base URL |
| `AUTH_ENABLED` | no | `false` | Enable Octo token and Space verification |
| `AUTH_CACHE_TTL` | no | `30s` | Successful identity cache duration |
| `AUTH_CACHE_CAPACITY` | no | `10000` | Maximum cached identities |
| `DEV_AUTH_UID` | no | `dev-user` | Local identity when auth is disabled |
| `DEV_AUTH_NAME` | no | `Developer` | Local display name when auth is disabled |
| `DEV_SPACE_ID` | no | `dev-space` | Local Space when auth is disabled |
| `HTTP_READ_HEADER_TIMEOUT` | no | `5s` | Header read timeout |
| `HTTP_READ_TIMEOUT` | no | `15s` | Request read timeout |
| `HTTP_WRITE_TIMEOUT` | no | `30s` | Response write timeout |
| `HTTP_IDLE_TIMEOUT` | no | `60s` | Keep-alive idle timeout |
| `SKIP_MIGRATION` | no | `false` | Skip embedded SQL migrations when `true` |

Example:

```bash
export MYSQL_DSN='marketplace:marketplace@tcp(127.0.0.1:3306)/octo_marketplace?charset=utf8mb4&parseTime=true'
go run ./cmd/marketplace-api
```

The credentials in `docker-compose.yaml` are development-only. Production must
provide rotated credentials through deployment-managed secrets.

## Authentication modes

Authentication is disabled by default so the service can run locally without
`octo-server`. In this mode, protected routes receive the configured development
identity and Space.

```bash
AUTH_ENABLED=false
DEV_AUTH_UID=dev-user
DEV_SPACE_ID=dev-space
```

Enable authentication when running with OCTO:

```bash
AUTH_ENABLED=true
OCTO_API_URL=http://octo-server:5001
```

Enabled mode validates tokens through
`POST /v1/auth/verify?include=context`, requires `X-Space-Id`, and verifies that
the authenticated user belongs to the requested Space. Production deployments
must set `AUTH_ENABLED=true`.
