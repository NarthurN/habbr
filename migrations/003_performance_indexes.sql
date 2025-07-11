-- Migration: 003_performance_indexes.sql
-- Description: Additional performance indexes and optimizations
-- Author: Auto-generated
-- Date: 2023

-- Partial indexes for specific use cases
CREATE INDEX idx_posts_enabled_comments_only
ON posts(created_at DESC)
WHERE comments_enabled = true;

CREATE INDEX idx_posts_by_author_recent
ON posts(author_id, created_at DESC)
WHERE created_at > (CURRENT_TIMESTAMP - INTERVAL '30 days');

-- Full-text search indexes (if needed)
CREATE INDEX idx_posts_title_search
ON posts USING gin(to_tsvector('english', title));

CREATE INDEX idx_posts_content_search
ON posts USING gin(to_tsvector('english', content));

CREATE INDEX idx_comments_content_search
ON comments USING gin(to_tsvector('english', content));

-- Index for hierarchical comment queries
CREATE INDEX idx_comments_hierarchical
ON comments(post_id, parent_id, depth, created_at);

-- Index for comment counting
CREATE INDEX idx_comments_count_by_post
ON comments(post_id)
WHERE parent_id IS NULL;

-- Statistics for query planner
ANALYZE posts;
ANALYZE comments;

-- Create materialized view for expensive aggregations (optional)
CREATE MATERIALIZED VIEW post_analytics AS
SELECT
    p.id,
    p.title,
    p.author_id,
    p.created_at,
    COUNT(c.id) as total_comments,
    COUNT(CASE WHEN c.parent_id IS NULL THEN 1 END) as root_comments,
    MAX(c.depth) as max_comment_depth,
    MAX(c.created_at) as last_comment_at,
    DATE_TRUNC('day', p.created_at) as created_date
FROM posts p
LEFT JOIN comments c ON p.id = c.post_id
GROUP BY p.id, p.title, p.author_id, p.created_at;

-- Index on materialized view
CREATE INDEX idx_post_analytics_author ON post_analytics(author_id);
CREATE INDEX idx_post_analytics_date ON post_analytics(created_date DESC);
CREATE INDEX idx_post_analytics_comments ON post_analytics(total_comments DESC);

-- Function to refresh materialized view
CREATE OR REPLACE FUNCTION refresh_post_analytics()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY post_analytics;
END;
$$ LANGUAGE plpgsql;

-- Function to get popular posts (example of performance optimization)
CREATE OR REPLACE FUNCTION get_popular_posts(
    limit_count INTEGER DEFAULT 10,
    days_back INTEGER DEFAULT 7
)
RETURNS TABLE(
    post_id UUID,
    title VARCHAR(200),
    author_id UUID,
    comment_count BIGINT,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        p.id,
        p.title,
        p.author_id,
        COUNT(c.id) as comment_count,
        p.created_at
    FROM posts p
    LEFT JOIN comments c ON p.id = c.post_id AND c.created_at > (CURRENT_TIMESTAMP - make_interval(days => days_back))
    WHERE p.created_at > (CURRENT_TIMESTAMP - make_interval(days => days_back))
    GROUP BY p.id, p.title, p.author_id, p.created_at
    ORDER BY comment_count DESC, p.created_at DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Function to get comment thread with proper ordering
CREATE OR REPLACE FUNCTION get_comment_thread(
    input_post_id UUID,
    max_depth INTEGER DEFAULT 10
)
RETURNS TABLE(
    id UUID,
    post_id UUID,
    parent_id UUID,
    content TEXT,
    author_id UUID,
    depth INTEGER,
    created_at TIMESTAMP WITH TIME ZONE,
    path TEXT
) AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE comment_tree AS (
        -- Base case: root comments
        SELECT
            c.id,
            c.post_id,
            c.parent_id,
            c.content,
            c.author_id,
            c.depth,
            c.created_at,
            c.id::TEXT as path
        FROM comments c
        WHERE c.post_id = input_post_id
        AND c.parent_id IS NULL

        UNION ALL

        -- Recursive case: child comments
        SELECT
            c.id,
            c.post_id,
            c.parent_id,
            c.content,
            c.author_id,
            c.depth,
            c.created_at,
            ct.path || '/' || c.id::TEXT as path
        FROM comments c
        INNER JOIN comment_tree ct ON c.parent_id = ct.id
        WHERE c.depth <= max_depth
    )
    SELECT
        ct.id,
        ct.post_id,
        ct.parent_id,
        ct.content,
        ct.author_id,
        ct.depth,
        ct.created_at,
        ct.path
    FROM comment_tree ct
    ORDER BY ct.path;
END;
$$ LANGUAGE plpgsql;
