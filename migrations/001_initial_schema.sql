-- Migration: 001_initial_schema.sql
-- Description: Initial schema for Habbr posts and comments system
-- Author: Auto-generated
-- Date: 2023

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create posts table
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL CHECK (length(trim(title)) > 0),
    content TEXT NOT NULL CHECK (length(trim(content)) > 0 AND length(content) <= 50000),
    author_id UUID NOT NULL,
    comments_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create comments table
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL CHECK (length(trim(content)) > 0 AND length(content) <= 2000),
    author_id UUID NOT NULL,
    depth INTEGER NOT NULL DEFAULT 0 CHECK (depth >= 0 AND depth <= 50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_posts_updated_at ON posts(updated_at DESC);
CREATE INDEX idx_posts_comments_enabled ON posts(comments_enabled);

CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
CREATE INDEX idx_comments_author_id ON comments(author_id);
CREATE INDEX idx_comments_created_at ON comments(created_at DESC);
CREATE INDEX idx_comments_updated_at ON comments(updated_at DESC);
CREATE INDEX idx_comments_depth ON comments(depth);

-- Composite indexes for common queries
CREATE INDEX idx_comments_post_parent ON comments(post_id, parent_id);
CREATE INDEX idx_comments_post_depth ON comments(post_id, depth);
CREATE INDEX idx_posts_author_created ON posts(author_id, created_at DESC);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers to automatically update updated_at
CREATE TRIGGER update_posts_updated_at
    BEFORE UPDATE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_comments_updated_at
    BEFORE UPDATE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create function to maintain comment depth
CREATE OR REPLACE FUNCTION calculate_comment_depth()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.depth = 0;
    ELSE
        SELECT depth + 1 INTO NEW.depth
        FROM comments
        WHERE id = NEW.parent_id;

        -- Validate depth limit
        IF NEW.depth > 50 THEN
            RAISE EXCEPTION 'Comment depth cannot exceed 50 levels';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically calculate comment depth
CREATE TRIGGER calculate_comments_depth
    BEFORE INSERT ON comments
    FOR EACH ROW
    EXECUTE FUNCTION calculate_comment_depth();

-- Create function to validate comment parent belongs to same post
CREATE OR REPLACE FUNCTION validate_comment_parent()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NOT NULL THEN
        PERFORM 1 FROM comments
        WHERE id = NEW.parent_id AND post_id = NEW.post_id;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'Parent comment must belong to the same post';
        END IF;
    END IF;

    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to validate comment parent
CREATE TRIGGER validate_comments_parent
    BEFORE INSERT OR UPDATE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION validate_comment_parent();

-- Create function to check if comments are enabled for post
CREATE OR REPLACE FUNCTION check_comments_enabled()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM 1 FROM posts
    WHERE id = NEW.post_id AND comments_enabled = true;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Comments are disabled for this post';
    END IF;

    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to check comments enabled (only for INSERT to allow updates/deletes)
CREATE TRIGGER check_comments_enabled_insert
    BEFORE INSERT ON comments
    FOR EACH ROW
    EXECUTE FUNCTION check_comments_enabled();

-- Create view for comment statistics
CREATE VIEW comment_stats AS
SELECT
    post_id,
    COUNT(*) as total_comments,
    MAX(depth) as max_depth,
    AVG(depth::numeric) as avg_depth,
    MAX(created_at) as last_comment_at
FROM comments
GROUP BY post_id;

-- Create view for post with comment counts
CREATE VIEW posts_with_stats AS
SELECT
    p.*,
    COALESCE(cs.total_comments, 0) as comment_count,
    COALESCE(cs.max_depth, 0) as max_comment_depth,
    cs.last_comment_at
FROM posts p
LEFT JOIN comment_stats cs ON p.id = cs.post_id;

-- Insert sample data (optional - for testing)
-- Uncomment the following lines if you want sample data

/*
INSERT INTO posts (id, title, content, author_id) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'Welcome to Habbr', 'This is the first post on our platform. Feel free to comment!', '550e8400-e29b-41d4-a716-446655440000'),
('550e8400-e29b-41d4-a716-446655440002', 'GraphQL is Awesome', 'Here''s why GraphQL is better than REST APIs...', '550e8400-e29b-41d4-a716-446655440000');

INSERT INTO comments (post_id, content, author_id) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'Great post! Looking forward to more content.', '550e8400-e29b-41d4-a716-446655440003'),
('550e8400-e29b-41d4-a716-446655440001', 'Thanks for sharing this information.', '550e8400-e29b-41d4-a716-446655440004');

-- Add a reply to first comment
INSERT INTO comments (post_id, parent_id, content, author_id)
SELECT
    '550e8400-e29b-41d4-a716-446655440001',
    id,
    'I totally agree with your comment!',
    '550e8400-e29b-41d4-a716-446655440000'
FROM comments
WHERE post_id = '550e8400-e29b-41d4-a716-446655440001'
AND parent_id IS NULL
LIMIT 1;
*/
