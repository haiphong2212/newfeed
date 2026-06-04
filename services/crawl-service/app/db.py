import asyncpg
import redis.asyncio as redis
from .config import settings


async def connect_db() -> asyncpg.Pool:
    return await asyncpg.create_pool(settings.database_url, min_size=1, max_size=5)


def connect_redis() -> redis.Redis:
    return redis.from_url(settings.redis_url, decode_responses=True)
