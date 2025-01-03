-- create extension and index for full text search

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_comments_content ON comments USING gin (content gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments (post_id);


CREATE INDEX IF NOT EXISTS idx_posts_title ON posts USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING gin (tags);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts (user_id);


CREATE INDEX IF NOT EXISTS idx_users_title ON users USING gin (username gin_trgm_ops);
