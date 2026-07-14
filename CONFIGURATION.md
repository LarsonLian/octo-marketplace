# Configuration

Environment variables configure the API service.

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `MYSQL_DSN` | yes | — | Marketplace MySQL DSN |
| `API_PORT` | no | `8092` | HTTP listen port |
| `OCTO_API_URL` | no | empty | Future `octo-server` API base URL |
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
