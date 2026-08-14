-- morph:nontransactional
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_fileinfo_name_lower ON fileinfo USING gin (LOWER(name) gin_bigm_ops);
