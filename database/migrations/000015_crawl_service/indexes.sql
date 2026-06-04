CREATE INDEX IF NOT EXISTS idx_crawl_sources_enabled ON crawl_sources(enabled);
CREATE INDEX IF NOT EXISTS idx_crawled_articles_source_id ON crawled_articles(source_id);
CREATE INDEX IF NOT EXISTS idx_crawled_articles_crawled_at ON crawled_articles(crawled_at DESC);
CREATE INDEX IF NOT EXISTS idx_generated_posts_status ON generated_posts(status);
CREATE INDEX IF NOT EXISTS idx_generated_posts_created_at ON generated_posts(created_at DESC);
