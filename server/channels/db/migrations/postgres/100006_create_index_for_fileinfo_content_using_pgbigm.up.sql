-- morph:nontransactional
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fileinfo_content_lower ON fileinfo USING gin (LOWER(content) gin_bigm_ops);
