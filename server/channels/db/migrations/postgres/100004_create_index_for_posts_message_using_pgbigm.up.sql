-- morph:nontransactional
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_posts_message_lower ON posts USING gin (LOWER(message) gin_bigm_ops);
