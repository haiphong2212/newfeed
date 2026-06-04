import hashlib
import json
import uuid
from datetime import datetime, timedelta, timezone
from typing import Any

import asyncpg
from redis.asyncio import Redis

from .config import settings


def fingerprint(value: str) -> str:
    return hashlib.sha256(value.encode("utf-8")).hexdigest()


async def list_sources(pool: asyncpg.Pool) -> list[dict[str, Any]]:
    rows = await pool.fetch("SELECT * FROM crawl_sources ORDER BY name")
    return [dict(row) for row in rows]


async def enabled_sources(pool: asyncpg.Pool) -> list[dict[str, Any]]:
    rows = await pool.fetch("SELECT * FROM crawl_sources WHERE enabled = TRUE ORDER BY name")
    return [dict(row) for row in rows]


async def create_source(pool: asyncpg.Pool, data: dict[str, Any]) -> dict[str, Any]:
    row = await pool.fetchrow(
        """
        INSERT INTO crawl_sources (name, base_url, rss_url, source_type, category, enabled, crawl_interval_minutes)
        VALUES ($1,$2,$3,$4,$5,$6,$7)
        ON CONFLICT (name) DO UPDATE SET
            base_url = EXCLUDED.base_url,
            rss_url = EXCLUDED.rss_url,
            source_type = EXCLUDED.source_type,
            category = EXCLUDED.category,
            enabled = EXCLUDED.enabled,
            crawl_interval_minutes = EXCLUDED.crawl_interval_minutes,
            updated_at = now()
        RETURNING *
        """,
        data["name"],
        data.get("base_url", ""),
        data.get("rss_url", ""),
        data.get("source_type", "rss"),
        data.get("category", "general"),
        data.get("enabled", True),
        data.get("crawl_interval_minutes", 60),
    )
    return dict(row)


async def save_crawled_article(pool: asyncpg.Pool, item: dict[str, Any]) -> dict[str, Any]:
    fp = fingerprint(item["url"])
    row = await pool.fetchrow(
        """
        INSERT INTO crawled_articles
            (source_id, original_url, canonical_url, title, author, image_url, raw_excerpt, extracted_text, fingerprint, published_at, status)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'cached')
        ON CONFLICT (original_url) DO UPDATE SET
            title = EXCLUDED.title,
            image_url = EXCLUDED.image_url,
            raw_excerpt = EXCLUDED.raw_excerpt,
            extracted_text = EXCLUDED.extracted_text,
            published_at = EXCLUDED.published_at,
            crawled_at = now(),
            updated_at = now()
        RETURNING *
        """,
        item.get("source_id"),
        item["url"],
        item.get("canonical_url", item["url"]),
        item["title"],
        item.get("author", ""),
        item.get("image_url", ""),
        item.get("excerpt", ""),
        item.get("text", ""),
        fp,
        item.get("published_at"),
    )
    return dict(row)


async def create_session(pool: asyncpg.Pool, query: str, requested_by: str | None, result_count: int) -> dict[str, Any]:
    cache_key = f"crawl:session:{uuid.uuid4()}"
    expires_at = datetime.now(timezone.utc) + timedelta(seconds=settings.crawl_cache_ttl_seconds)
    row = await pool.fetchrow(
        """
        INSERT INTO crawl_search_sessions (query, requested_by, cache_key, result_count, expires_at)
        VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5)
        RETURNING *
        """,
        query,
        requested_by or "",
        cache_key,
        result_count,
        expires_at,
    )
    return dict(row)


async def cache_candidates(redis: Redis, cache_key: str, candidates: list[dict[str, Any]]) -> None:
    payload = json.dumps(candidates, default=str, ensure_ascii=False)
    await redis.set(cache_key, payload, ex=settings.crawl_cache_ttl_seconds)
    for candidate in candidates:
        await redis.set(f"crawl:candidate:{candidate['id']}", payload, ex=settings.crawl_cache_ttl_seconds)


async def load_candidate(redis: Redis, candidate_id: str) -> tuple[dict[str, Any], list[dict[str, Any]], str] | None:
    keys = await redis.keys("crawl:session:*")
    for key in keys:
        raw = await redis.get(key)
        if not raw:
            continue
        candidates = json.loads(raw)
        for candidate in candidates:
            if candidate["id"] == candidate_id:
                return candidate, candidates, key
    return None


async def approve_candidate(pool: asyncpg.Pool, candidate: dict[str, Any], approved_by: str | None) -> dict[str, Any]:
    source_published_at = candidate.get("source_published_at")
    if isinstance(source_published_at, str) and source_published_at:
        source_published_at = datetime.fromisoformat(source_published_at.replace("Z", "+00:00"))
    row = await pool.fetchrow(
        """
        INSERT INTO generated_posts
            (crawled_article_id, approved_by, title, short_content, source_name, source_url, image_url, category, tags,
             status, source_published_at, approved_at, published_at)
        VALUES ($1::uuid, NULLIF($2, '')::uuid, $3,$4,$5,$6,$7,$8,$9,'approved',$10,now(),now())
        RETURNING *
        """,
        candidate["crawled_article_id"],
        approved_by or "",
        candidate["title"],
        candidate["short_content"],
        candidate["source_name"],
        candidate["source_url"],
        candidate.get("image_url", ""),
        candidate.get("category", "general"),
        candidate.get("tags", []),
        source_published_at,
    )
    await pool.execute("UPDATE crawled_articles SET status = 'approved', updated_at = now() WHERE id = $1::uuid", candidate["crawled_article_id"])
    return dict(row)


async def cleanup_session(redis: Redis, cache_key: str) -> None:
    await redis.delete(cache_key)


async def delete_unselected_articles(pool: asyncpg.Pool, candidate: dict[str, Any], candidates: list[dict[str, Any]]) -> None:
    for other in candidates:
        if other["id"] != candidate["id"]:
            await pool.execute("DELETE FROM crawled_articles WHERE id = $1::uuid AND status = 'cached'", other["crawled_article_id"])


async def delete_session(pool: asyncpg.Pool, cache_key: str) -> None:
    await pool.execute("DELETE FROM crawl_search_sessions WHERE cache_key = $1", cache_key)


async def list_posts(pool: asyncpg.Pool, limit: int = 20) -> list[dict[str, Any]]:
    rows = await pool.fetch("SELECT * FROM generated_posts ORDER BY created_at DESC LIMIT $1", limit)
    return [dict(row) for row in rows]
