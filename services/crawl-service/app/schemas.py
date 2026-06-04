from datetime import datetime
from typing import List
from pydantic import BaseModel, Field, HttpUrl


class CrawlSourceIn(BaseModel):
    name: str
    base_url: str = ""
    rss_url: str = ""
    source_type: str = "rss"
    category: str = "general"
    enabled: bool = True
    crawl_interval_minutes: int = 60


class SearchRequest(BaseModel):
    query: str = Field(min_length=2)
    limit: int = 5


class Candidate(BaseModel):
    id: str
    session_id: str
    crawled_article_id: str
    title: str
    short_content: str
    source_name: str
    source_url: str
    image_url: str = ""
    category: str = "general"
    tags: List[str] = []
    source_published_at: datetime | None = None
    created_at: datetime


class SearchResponse(BaseModel):
    session_id: str
    candidates: List[Candidate]
    expires_at: datetime


class ApproveResponse(BaseModel):
    post_id: str
    status: str
