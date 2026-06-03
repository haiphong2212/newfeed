CREATE INDEX IF NOT EXISTS idx_article_comments_article_created_at ON article_comments(article_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_article_comments_user_created_at ON article_comments(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_article_shares_user_created_at ON article_shares(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_article_shares_article_created_at ON article_shares(article_id, created_at DESC);
