# Media Storage

`media-service` supports two object storage drivers.

## Local

Use local storage for development or a single VPS:

```env
OBJECT_STORAGE_DRIVER=local
OBJECT_STORAGE_ROOT=/var/lib/newfeed/objects
MEDIA_BUCKET=community-news
MEDIA_PUBLIC_URL=http://localhost:8008
```

Docker Compose mounts this path through the `object_storage` volume.

## Cloudflare R2

Use R2 for production object storage:

```env
OBJECT_STORAGE_DRIVER=r2
R2_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com
R2_ACCESS_KEY_ID=<access-key-id>
R2_SECRET_ACCESS_KEY=<secret-access-key>
R2_BUCKET=<bucket-name>
R2_PUBLIC_URL=<public-bucket-url-or-custom-domain>
```

The upload endpoint remains:

```http
POST /v1/media/upload
```

Multipart fields:

- `file`
- `owner_id`
- `bucket` optional; defaults to `MEDIA_BUCKET`

PostgreSQL stores object metadata in `media_objects`; R2 stores the object bytes.

For uploads, the R2 S3 credentials must have object write permission. Read/list-only credentials can list or read existing objects but cannot upload new media.
