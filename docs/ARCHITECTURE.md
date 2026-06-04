# Architecture

This repository is organized as a Go monorepo with independent service entrypoints under `services/*/cmd`.

Each service follows feature-first clean architecture:

- `internal/<feature>/domain`: business entities, value objects, and business rules.
- `internal/<feature>/usecase`: application workflows and ports/interfaces.
- `internal/<feature>/repository`: persistence adapters.
- `internal/<feature>/delivery`: HTTP, gRPC, or WebSocket adapters.
- `internal/platform`: service-local infrastructure such as config, security, events, logging, database clients.

Current implementation path:

- `auth-service` has working register, login, refresh-token rotation, and access-token validation.
- `auth-service` stores users and refresh tokens in PostgreSQL.
- `news-service` stores articles in PostgreSQL, caches article reads in Redis, and emits `ArticlePublished` through RabbitMQ.
- `search-service` indexes and queries article documents in Elasticsearch.
- `analytics-service` writes trending scores to the Redis sorted set `trending_articles`.
- `media-service` stores object bytes on the server filesystem and metadata in PostgreSQL.
- Every service runs Fiber for HTTP and a gRPC health server on its internal port.

The ports remain in the usecase layer, so infrastructure adapters can evolve without moving business logic into delivery or persistence code.
