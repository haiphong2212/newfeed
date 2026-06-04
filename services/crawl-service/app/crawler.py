import re
import uuid
from datetime import datetime, timezone
from email.utils import parsedate_to_datetime
from typing import Any

import feedparser
import httpx
import trafilatura
from bs4 import BeautifulSoup

from .config import settings


def compact(text: str) -> str:
    return re.sub(r"\s+", " ", text or "").strip()


def summarize(text: str, limit: int = 420) -> str:
    text = compact(text)
    if len(text) <= limit:
        return text
    sentences = re.split(r"(?<=[.!?。])\s+", text)
    out = ""
    for sentence in sentences:
        if len(out) + len(sentence) + 1 > limit:
            break
        out = compact(out + " " + sentence)
    return out or text[:limit].rsplit(" ", 1)[0]


def tags_from_text(query: str, title: str, category: str) -> list[str]:
    words = re.findall(r"[\wÀ-ỹ]+", f"{query} {title}".lower())
    ignored = {"the", "and", "with", "cho", "cua", "của", "bao", "báo", "tin", "moi", "mới"}
    tags: list[str] = []
    for word in words:
        if len(word) < 3 or word in ignored or word in tags:
            continue
        tags.append(word)
        if len(tags) >= 4:
            break
    if category and category not in tags:
        tags.append(category)
    return tags


def parse_datetime(value: str | None) -> datetime | None:
    if not value:
        return None
    try:
        parsed = parsedate_to_datetime(value)
        if parsed.tzinfo is None:
            return parsed.replace(tzinfo=timezone.utc)
        return parsed.astimezone(timezone.utc)
    except Exception:
        return None


def entry_image(entry: Any) -> str:
    media = getattr(entry, "media_content", None) or []
    if media and media[0].get("url"):
        return media[0]["url"]
    links = getattr(entry, "links", None) or []
    for link in links:
        if str(link.get("type", "")).startswith("image/") and link.get("href"):
            return link["href"]
    summary = getattr(entry, "summary", "")
    soup = BeautifulSoup(summary, "html.parser")
    img = soup.find("img")
    return img.get("src", "") if img else ""


async def fetch_article_text(url: str) -> tuple[str, str]:
    async with httpx.AsyncClient(timeout=settings.crawl_timeout_seconds, follow_redirects=True) as client:
        res = await client.get(url, headers={"User-Agent": "NewFeedCrawler/1.0"})
        res.raise_for_status()
    downloaded = res.text
    extracted = trafilatura.extract(downloaded, include_comments=False, include_tables=False) or ""
    soup = BeautifulSoup(downloaded, "html.parser")
    image = ""
    og = soup.find("meta", property="og:image")
    if og and og.get("content"):
        image = og["content"]
    return compact(extracted), image


async def crawl_candidates(query: str, sources: list[dict[str, Any]], limit: int) -> list[dict[str, Any]]:
    query_norm = query.lower().strip()
    matches: list[dict[str, Any]] = []
    async with httpx.AsyncClient(timeout=settings.crawl_timeout_seconds, follow_redirects=True) as client:
        for source in sources:
            rss_url = source.get("rss_url") or ""
            if not rss_url:
                continue
            try:
                res = await client.get(rss_url, headers={"User-Agent": "NewFeedCrawler/1.0"})
                res.raise_for_status()
            except Exception:
                continue
            feed = feedparser.parse(res.text)
            for entry in feed.entries[:30]:
                title = compact(getattr(entry, "title", ""))
                summary = compact(BeautifulSoup(getattr(entry, "summary", ""), "html.parser").get_text(" "))
                url = getattr(entry, "link", "")
                haystack = f"{title} {summary}".lower()
                if query_norm not in haystack and all(part not in haystack for part in query_norm.split()):
                    continue
                text, article_image = "", ""
                if url:
                    try:
                        text, article_image = await fetch_article_text(url)
                    except Exception:
                        text = summary
                image_url = article_image or entry_image(entry)
                content_source = text or summary or title
                category = source.get("category") or "general"
                matches.append(
                    {
                        "id": str(uuid.uuid4()),
                        "source_id": str(source["id"]),
                        "source_name": source["name"],
                        "url": url,
                        "canonical_url": url,
                        "title": title,
                        "author": getattr(entry, "author", ""),
                        "image_url": image_url,
                        "excerpt": summary,
                        "text": content_source,
                        "short_content": summarize(content_source),
                        "category": category,
                        "tags": tags_from_text(query, title, category),
                        "published_at": parse_datetime(getattr(entry, "published", None)),
                        "created_at": datetime.now(timezone.utc),
                    }
                )
                if len(matches) >= limit:
                    return matches
    return matches[:limit]
