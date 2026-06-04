from fastapi import Header, HTTPException, Request
import httpx
import asyncpg
from .config import settings


async def require_admin(
    request: Request,
    authorization: str | None = Header(default=None),
    x_admin_bypass: str | None = Header(default=None),
) -> dict:
    if settings.admin_bypass_token and x_admin_bypass == settings.admin_bypass_token:
        return {"user_id": None, "role": "admin", "email": "local-admin"}
    if not authorization:
        raise HTTPException(status_code=401, detail="authorization header is required")
    async with httpx.AsyncClient(timeout=5) as client:
        res = await client.get(settings.auth_validate_url, headers={"Authorization": authorization})
    if res.status_code >= 400:
        raise HTTPException(status_code=401, detail="invalid token")
    claims = res.json()
    user_id = claims.get("user_id") or claims.get("UserID")
    role = claims.get("role") or claims.get("Role")
    if role not in {"admin", "editor"}:
        pool: asyncpg.Pool = request.app.state.db
        allowed = await pool.fetchval(
            """
            SELECT EXISTS (
                SELECT 1
                FROM user_roles ur
                JOIN roles r ON r.id = ur.role_id
                WHERE ur.user_id = $1::uuid AND r.name IN ('admin', 'editor')
            )
            """,
            user_id,
        )
        if not allowed:
            raise HTTPException(status_code=403, detail="admin or editor role is required")
    return {"user_id": user_id, "role": role, "email": claims.get("email") or claims.get("Email")}
