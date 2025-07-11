-- Migration: 002_rollback_initial.sql
-- Description: Rollback migration for initial schema
-- Author: Auto-generated
-- Date: 2023

-- Drop views
DROP VIEW IF EXISTS posts_with_stats;
DROP VIEW IF EXISTS comment_stats;

-- Drop triggers
DROP TRIGGER IF EXISTS check_comments_enabled_insert ON comments;
DROP TRIGGER IF EXISTS validate_comments_parent ON comments;
DROP TRIGGER IF EXISTS calculate_comments_depth ON comments;
DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;
DROP TRIGGER IF EXISTS update_posts_updated_at ON posts;

-- Drop functions
DROP FUNCTION IF EXISTS check_comments_enabled();
DROP FUNCTION IF EXISTS validate_comment_parent();
DROP FUNCTION IF EXISTS calculate_comment_depth();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_posts_author_created;
DROP INDEX IF EXISTS idx_comments_post_depth;
DROP INDEX IF EXISTS idx_comments_post_parent;
DROP INDEX IF EXISTS idx_comments_depth;
DROP INDEX IF EXISTS idx_comments_updated_at;
DROP INDEX IF EXISTS idx_comments_created_at;
DROP INDEX IF EXISTS idx_comments_author_id;
DROP INDEX IF EXISTS idx_comments_parent_id;
DROP INDEX IF EXISTS idx_comments_post_id;
DROP INDEX IF EXISTS idx_posts_comments_enabled;
DROP INDEX IF EXISTS idx_posts_updated_at;
DROP INDEX IF EXISTS idx_posts_created_at;
DROP INDEX IF EXISTS idx_posts_author_id;

-- Drop tables (comments first due to foreign key constraints)
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;

-- Drop extension (only if no other objects use it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
