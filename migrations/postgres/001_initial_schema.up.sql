-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create URLs table
CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    original_url TEXT NOT NULL,
    hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create short_links table
CREATE TABLE IF NOT EXISTS short_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code TEXT NOT NULL,
    custom_alias TEXT UNIQUE,
    url_id UUID NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    expiration_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create link_clicks table
CREATE TABLE IF NOT EXISTS link_clicks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    short_link_id UUID NOT NULL REFERENCES short_links(id) ON DELETE CASCADE,
    referrer TEXT,
    user_agent TEXT,
    ip_address TEXT,
    country TEXT,
    city TEXT,
    device TEXT,
    browser TEXT,
    os TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_urls_hash ON urls(hash);
CREATE INDEX IF NOT EXISTS idx_short_links_code ON short_links(code);
CREATE INDEX IF NOT EXISTS idx_short_links_custom_alias ON short_links(custom_alias);
CREATE INDEX IF NOT EXISTS idx_short_links_url_id ON short_links(url_id);
CREATE INDEX IF NOT EXISTS idx_short_links_is_active ON short_links(is_active);
CREATE INDEX IF NOT EXISTS idx_short_links_expiration_date ON short_links(expiration_date);
CREATE INDEX IF NOT EXISTS idx_link_clicks_short_link_id ON link_clicks(short_link_id);
CREATE INDEX IF NOT EXISTS idx_link_clicks_created_at ON link_clicks(created_at); 