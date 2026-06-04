CREATE TABLE IF NOT EXISTS crawl_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    base_url TEXT NOT NULL DEFAULT '',
    rss_url TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT 'rss',
    category TEXT NOT NULL DEFAULT 'general',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    crawl_interval_minutes INT NOT NULL DEFAULT 60,
    last_crawled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS crawled_articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID REFERENCES crawl_sources(id) ON DELETE SET NULL,
    original_url TEXT NOT NULL UNIQUE,
    canonical_url TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT '',
    image_url TEXT NOT NULL DEFAULT '',
    raw_excerpt TEXT NOT NULL DEFAULT '',
    extracted_text TEXT NOT NULL DEFAULT '',
    fingerprint TEXT NOT NULL UNIQUE,
    published_at TIMESTAMPTZ,
    crawled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    status TEXT NOT NULL DEFAULT 'crawled',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS crawl_search_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    query TEXT NOT NULL,
    requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    cache_key TEXT NOT NULL UNIQUE,
    result_count INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS generated_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    crawled_article_id UUID REFERENCES crawled_articles(id) ON DELETE SET NULL,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    short_content TEXT NOT NULL,
    source_name TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL,
    image_url TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'general',
    tags TEXT[] NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'draft',
    source_published_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO crawl_sources (name, base_url, rss_url, category)
VALUES
    ('VnExpress Technology', 'https://vnexpress.net', 'https://vnexpress.net/rss/so-hoa.rss', 'technology'),
    ('Tuoi Tre Technology', 'https://tuoitre.vn', 'https://tuoitre.vn/rss/nhip-song-so.rss', 'technology'),
    ('Thanh Nien Technology', 'https://thanhnien.vn', 'https://thanhnien.vn/rss/cong-nghe.rss', 'technology')
ON CONFLICT (name) DO NOTHING;
