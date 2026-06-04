import asyncio
from contextlib import asynccontextmanager
from datetime import datetime, timezone

import uvicorn
from fastapi import Depends, FastAPI, HTTPException, Request

from . import repository
from .auth import require_admin
from .config import settings
from .crawler import crawl_candidates
from .db import connect_db, connect_redis
from .events import publish_event
from .grpc_server import serve_grpc
from .schemas import ApproveResponse, Candidate, CrawlSourceIn, SearchRequest, SearchResponse


@asynccontextmanager
async def lifespan(app: FastAPI):
    app.state.db = await connect_db()
    app.state.redis = connect_redis()
    app.state.grpc_stop = asyncio.Event()
    app.state.grpc_task = asyncio.create_task(serve_grpc(app.state.grpc_stop))
    yield
    app.state.grpc_stop.set()
    await app.state.grpc_task
    await app.state.redis.close()
    await app.state.db.close()


app = FastAPI(title="NewFeed Crawl Service", version="1.0.0", lifespan=lifespan)


@app.get("/healthz")
async def healthz():
    return {"service": settings.service_name, "status": "ok"}


@app.get("/v1/admin/crawl/sources")
async def get_sources(request: Request, _admin=Depends(require_admin)):
    return {"sources": await repository.list_sources(request.app.state.db)}


@app.post("/v1/admin/crawl/sources")
async def post_source(payload: CrawlSourceIn, request: Request, _admin=Depends(require_admin)):
    return await repository.create_source(request.app.state.db, payload.model_dump())


@app.post("/v1/admin/crawl/search", response_model=SearchResponse)
async def search(payload: SearchRequest, request: Request, admin=Depends(require_admin)):
    limit = min(max(payload.limit, 3), settings.max_candidates)
    sources = await repository.enabled_sources(request.app.state.db)
    raw_candidates = await crawl_candidates(payload.query, sources, limit)
    candidates: list[dict] = []
    for raw in raw_candidates:
        crawled = await repository.save_crawled_article(request.app.state.db, raw)
        candidates.append(
            {
                "id": raw["id"],
                "session_id": "",
                "crawled_article_id": str(crawled["id"]),
                "title": raw["title"],
                "short_content": raw["short_content"],
                "source_name": raw["source_name"],
                "source_url": raw["url"],
                "image_url": raw.get("image_url", ""),
                "category": raw.get("category", "general"),
                "tags": raw.get("tags", []),
                "source_published_at": raw.get("published_at"),
                "created_at": datetime.now(timezone.utc),
            }
        )
    session = await repository.create_session(request.app.state.db, payload.query, admin.get("user_id"), len(candidates))
    for item in candidates:
        item["session_id"] = str(session["id"])
    await repository.cache_candidates(request.app.state.redis, session["cache_key"], candidates)
    return {"session_id": str(session["id"]), "candidates": candidates, "expires_at": session["expires_at"]}


@app.post("/v1/admin/crawl/candidates/{candidate_id}/approve", response_model=ApproveResponse)
async def approve(candidate_id: str, request: Request, admin=Depends(require_admin)):
    loaded = await repository.load_candidate(request.app.state.redis, candidate_id)
    if not loaded:
        raise HTTPException(status_code=404, detail="candidate expired or not found")
    candidate, candidates, cache_key = loaded
    post = await repository.approve_candidate(request.app.state.db, candidate, admin.get("user_id"))
    await repository.delete_unselected_articles(request.app.state.db, candidate, candidates)
    await repository.cleanup_session(request.app.state.redis, cache_key)
    await repository.delete_session(request.app.state.db, cache_key)
    await publish_event(
        "CrawlPostApproved",
        {"post_id": str(post["id"]), "title": post["title"], "source_url": post["source_url"], "image_url": post["image_url"]},
    )
    return {"post_id": str(post["id"]), "status": post["status"]}


@app.get("/v1/admin/crawl/posts")
async def posts(request: Request, _admin=Depends(require_admin), limit: int = 20):
    return {"posts": await repository.list_posts(request.app.state.db, min(limit, 100))}


if __name__ == "__main__":
    uvicorn.run("app.main:app", host=settings.http_host, port=settings.http_port)
