-- Drop indexes
DROP INDEX IF EXISTS idx_link_clicks_created_at;
DROP INDEX IF EXISTS idx_link_clicks_short_link_id;
DROP INDEX IF EXISTS idx_short_links_expiration_date;
DROP INDEX IF EXISTS idx_short_links_is_active;
DROP INDEX IF EXISTS idx_short_links_url_id;
DROP INDEX IF EXISTS idx_short_links_custom_alias;
DROP INDEX IF EXISTS idx_short_links_code;
DROP INDEX IF EXISTS idx_urls_hash;

-- Drop tables
DROP TABLE IF EXISTS link_clicks;
DROP TABLE IF EXISTS short_links;
DROP TABLE IF EXISTS urls; 