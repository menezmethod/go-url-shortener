-- Rollback script for initial schema

-- Drop link_clicks table and its indexes
DROP TABLE IF EXISTS link_clicks CASCADE;

-- Drop short_links table and its indexes
DROP TABLE IF EXISTS short_links CASCADE;

-- Drop urls table and its indexes
DROP TABLE IF EXISTS urls CASCADE; 