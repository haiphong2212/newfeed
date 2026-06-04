# Cyber Security

## Authentication

- Access tokens are JWTs signed with HMAC-SHA256.
- Tokens include `sub`, `email`, `role`, `iss`, `aud`, `iat`, `nbf`, and `exp`.
- The JWT secret must be at least 32 bytes.
- Passwords are hashed with bcrypt.
- Login and refresh responses can set secure cookies:
  - `HttpOnly`
  - `SameSite`
  - optional `Secure`
  - optional cookie domain

For production:

```env
AUTH_COOKIE_ENABLED=true
AUTH_COOKIE_SECURE=true
AUTH_COOKIE_SAME_SITE=Strict
JWT_SECRET=<long-random-secret-at-least-32-bytes>
```

Use `SameSite=Lax` if the frontend and API are same-site but on different subdomains. Use `SameSite=None` only when cross-site cookies are required, and then `AUTH_COOKIE_SECURE=true` is mandatory.

## HTTP hardening

All Fiber apps use security headers:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection`
- strict referrer policy
- cross-origin resource policies

## Media upload controls

`media-service` accepts multipart uploads through:

```http
POST /v1/media/upload
```

Accepted fields:

- `files`: multiple files
- `file`: single file compatibility

Limits:

- maximum 3 image files per request
- video files must be 20MB or smaller
- only `image/*` and `video/*` content types are accepted
- default request body limit is 64MB through `HTTP_BODY_LIMIT_BYTES`

## Remaining production hardening

- Add CSRF tokens if browser clients rely on cookies for authenticated state-changing requests.
- Add per-user and per-IP rate limiting at API Gateway.
- Add malware scanning for uploaded media before public serving.
- Rotate leaked or exposed credentials immediately.
