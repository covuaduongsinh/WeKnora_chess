-- Down migration for 000001_wiki_and_indexing (SQLite / Lite)
DROP TABLE IF EXISTS wiki_page_issues;
DROP TABLE IF EXISTS wiki_folders;
DROP TABLE IF EXISTS wiki_pages;

-- DROP COLUMN requires SQLite >= 3.35 (bundled with the modern go-sqlite3
-- amalgamation used by the Lite build).
ALTER TABLE knowledge_bases DROP COLUMN indexing_strategy;
ALTER TABLE knowledge_bases DROP COLUMN wiki_config;
