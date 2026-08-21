-- Rollback: 000910_chess_articles
DROP INDEX IF EXISTS idx_chess_slug_aliases_kind;
ALTER TABLE chess_slug_aliases DROP COLUMN IF EXISTS kind;

DROP INDEX IF EXISTS idx_chess_article_revisions_article;
DROP TABLE IF EXISTS chess_article_revisions;

DROP INDEX IF EXISTS idx_chess_article_images_article;
DROP TABLE IF EXISTS chess_article_images;

DROP INDEX IF EXISTS idx_chess_article_topic_items_topic;
DROP INDEX IF EXISTS idx_chess_article_topic_items_article;
DROP TABLE IF EXISTS chess_article_topic_items;

DROP INDEX IF EXISTS idx_chess_article_topics_parent;
DROP INDEX IF EXISTS idx_chess_article_topics_tenant_slug;
DROP TABLE IF EXISTS chess_article_topics;

DROP INDEX IF EXISTS idx_chess_articles_title_trgm;
DROP INDEX IF EXISTS idx_chess_articles_fulltext;
DROP INDEX IF EXISTS idx_chess_articles_status;
DROP INDEX IF EXISTS idx_chess_articles_filter;
DROP INDEX IF EXISTS idx_chess_articles_tenant;
DROP INDEX IF EXISTS idx_chess_articles_tenant_slug;
DROP TABLE IF EXISTS chess_articles;
