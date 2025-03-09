-- Migrations for URL Shortener Database

-- Create URLs table
CREATE TABLE urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_url TEXT NOT NULL,
    hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on hash for faster lookups
CREATE UNIQUE INDEX idx_urls_hash ON urls(hash);

-- Create short_links table
CREATE TABLE short_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL,
    custom_alias TEXT UNIQUE,
    url_id UUID NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
    expiration_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for short_links
CREATE UNIQUE INDEX idx_short_links_code ON short_links(code);
CREATE INDEX idx_short_links_custom_alias ON short_links(custom_alias) WHERE custom_alias IS NOT NULL;
CREATE INDEX idx_short_links_url_id ON short_links(url_id);
CREATE INDEX idx_short_links_expiration ON short_links(expiration_date) WHERE expiration_date IS NOT NULL;
CREATE INDEX idx_short_links_is_active ON short_links(is_active);

-- Create link_clicks table for analytics
CREATE TABLE link_clicks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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

-- Create indexes for link_clicks
CREATE INDEX idx_link_clicks_short_link_id ON link_clicks(short_link_id);
CREATE INDEX idx_link_clicks_created_at ON link_clicks(created_at); 