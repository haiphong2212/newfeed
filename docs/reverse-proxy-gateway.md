# Reverse Proxy Gateway

The gateway is a standard-library Go reverse proxy that can replace Nginx in front of NewFeed.

## Routing

- `web.newfeed.site` -> `APP_TARGET`
- `admin.newfeed.site` -> `ADMIN_TARGET`
- `api.newfeed.site` -> `API_TARGET`
- `web.newfeed.site/api/*` -> `API_TARGET` with `/api` stripped
- `web.newfeed.site/ws/*` -> `CHAT_TARGET`

Defaults:

```env
APP_TARGET=http://frontend-user:3000
ADMIN_TARGET=http://frontend-admin:3001
API_TARGET=http://api:8080
CHAT_TARGET=http://chat-service:8006
```

## Run Locally

```bash
go run ./cmd/gateway
```

## Docker Compose Production

```bash
docker compose -f deployments/docker/docker-compose.gateway.yml up --build
```

Only the `gateway` service publishes host ports. Frontend, API, PostgreSQL, and Redis stay on the internal Docker network.

## CSRF

When browser clients authenticate with cookies, call:

```http
GET /csrf-token
```

The gateway sets a `csrf_token` cookie and returns the same token in JSON. For `POST`, `PUT`, `PATCH`, and `DELETE`, send:

```http
X-CSRF-Token: <token>
```

The CSRF check is enforced only when one of the configured auth cookies is present.
