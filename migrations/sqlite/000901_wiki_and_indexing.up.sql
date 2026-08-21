-- Migration: 000001_wiki_and_indexing (SQLite / Lite)
-- Mirrors the Postgres migration 000037_wiki_and_indexing for the SQLite build:
--   * adds wiki_config + indexing_strategy columns to knowledge_bases
--   * creates wiki_pages, wiki_folders, wiki_page_issues tables
-- SQLite has no JSONB / GIN / "TIMESTAMP WITH TIME ZONE": JSON columns are TEXT
-- (GORM (de)serialises via the Valuer/Scanner methods on the Go types) and
-- timestamps are DATETIME. The Postgres-only fulltext GIN index is intentionally
-- omitted — the wiki repository's text-search branch (to_tsvector/plainto_tsquery)
-- is Postgres-only, so SQLite needs no equivalent index.

-- ---------------------------------------------------------------------------
-- 1) knowledge_bases.wiki_config + indexing_strategy
-- SQLite cannot ADD COLUMN IF NOT EXISTS; these columns are absent from the
-- consolidated init schema, so a plain ADD COLUMN is safe.
-- ---------------------------------------------------------------------------
ALTER TABLE knowledge_bases ADD COLUMN wiki_config TEXT;
ALTER TABLE knowledge_bases ADD COLUMN indexing_strategy TEXT;

-- Backfill existing rows with the legacy default (vector + keyword on, wiki +
-- graph off) so older knowledge bases keep their pre-wiki behavior.
UPDATE knowledge_bases
SET indexing_strategy = '{"vector_enabled":true,"keyword_enabled":true,"wiki_enabled":false,"graph_enabled":false}'
WHERE indexing_strategy IS NULL;

-- ---------------------------------------------------------------------------
-- 2) wiki_pages
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS wiki_pages (
    id                VARCHAR(36) PRIMARY KEY,
    tenant_id         INTEGER NOT NULL,
    knowledge_base_id VARCHAR(36) NOT NULL,
    slug              VARCHAR(255) NOT NULL,
    title             VARCHAR(512) NOT NULL DEFAULT '',
    page_type         VARCHAR(32) NOT NULL DEFAULT 'summary',
    status            VARCHAR(32) NOT NULL DEFAULT 'published',
    content           TEXT NOT NULL DEFAULT '',
    summary           TEXT NOT NULL DEFAULT '',
    parent_slug       VARCHAR(255) NOT NULL DEFAULT '',
    folder_id         VARCHAR(36) NOT NULL DEFAULT '',
    category_path     TEXT DEFAULT '[]',
    wiki_path         VARCHAR(1024) NOT NULL DEFAULT '',
    depth             INTEGER NOT NULL DEFAULT 0,
    sort_order        INTEGER NOT NULL DEFAULT 0,
    source_refs       TEXT DEFAULT '[]',
    chunk_refs        TEXT DEFAULT '[]',
    in_links          TEXT DEFAULT '[]',
    out_links         TEXT DEFAULT '[]',
    page_metadata     TEXT DEFAULT '{}',
    aliases           TEXT DEFAULT '[]',
    version           INTEGER NOT NULL DEFAULT 1,
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at        DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wiki_pages_kb_slug
    ON wiki_pages (knowledge_base_id, slug)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_wiki_pages_kb_id
    ON wiki_pages (knowledge_base_id);
CREATE INDEX IF NOT EXISTS idx_wiki_pages_page_type
    ON wiki_pages (knowledge_base_id, page_type);
CREATE INDEX IF NOT EXISTS idx_wiki_pages_parent_slug
    ON wiki_pages (knowledge_base_id, parent_slug);
CREATE INDEX IF NOT EXISTS idx_wiki_pages_tree
    ON wiki_pages (knowledge_base_id, page_type, wiki_path, sort_order, title);
CREATE INDEX IF NOT EXISTS idx_wiki_pages_folder
    ON wiki_pages (knowledge_base_id, folder_id);
CREATE INDEX IF NOT EXISTS idx_wiki_pages_tenant_id
    ON wiki_pages (tenant_id);
CREATE INDEX IF NOT EXISTS idx_wiki_pages_deleted_at
    ON wiki_pages (deleted_at);

-- ---------------------------------------------------------------------------
-- 3) wiki_folders
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS wiki_folders (
    id                VARCHAR(36) PRIMARY KEY,
    tenant_id         INTEGER NOT NULL DEFAULT 0,
    knowledge_base_id VARCHAR(36) NOT NULL,
    parent_id         VARCHAR(36) NOT NULL DEFAULT '',
    name              VARCHAR(255) NOT NULL,
    path              VARCHAR(1024) NOT NULL DEFAULT '',
    depth             INTEGER NOT NULL DEFAULT 0,
    sort_order        INTEGER NOT NULL DEFAULT 0,
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at        DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wiki_folders_parent_name
    ON wiki_folders (knowledge_base_id, parent_id, name)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_wiki_folders_parent
    ON wiki_folders (knowledge_base_id, parent_id);
CREATE INDEX IF NOT EXISTS idx_wiki_folders_deleted_at
    ON wiki_folders (deleted_at);

-- ---------------------------------------------------------------------------
-- 4) wiki_page_issues
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS wiki_page_issues (
    id                      VARCHAR(36) PRIMARY KEY,
    tenant_id               INTEGER NOT NULL,
    knowledge_base_id       VARCHAR(36) NOT NULL,
    slug                    VARCHAR(255) NOT NULL,
    issue_type              VARCHAR(50) NOT NULL,
    description             TEXT NOT NULL,
    suspected_knowledge_ids TEXT,
    status                  VARCHAR(20) NOT NULL DEFAULT 'pending',
    reported_by             VARCHAR(100) NOT NULL,
    created_at              DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at              DATETIME
);

CREATE INDEX IF NOT EXISTS idx_wiki_page_issues_tenant_id ON wiki_page_issues(tenant_id);
CREATE INDEX IF NOT EXISTS idx_wiki_page_issues_knowledge_base_id ON wiki_page_issues(knowledge_base_id);
CREATE INDEX IF NOT EXISTS idx_wiki_page_issues_slug ON wiki_page_issues(slug);
CREATE INDEX IF NOT EXISTS idx_wiki_page_issues_status ON wiki_page_issues(status);
