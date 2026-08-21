-- Migration: 000002_lite_missing_tables (SQLite / Lite)
-- Backports tables that exist in the Postgres versioned migrations but were
-- never added to the consolidated SQLite init schema. Without task_pending_ops
-- the wiki-ingest subtask cannot enqueue its pending op, so a document with the
-- Wiki Knowledge Base indexing enabled hangs forever in "finalizing".
-- Type mapping: BIGSERIAL -> INTEGER PRIMARY KEY AUTOINCREMENT, BIGINT -> INTEGER,
-- JSONB -> TEXT, TIMESTAMP[TZ] -> DATETIME, NOW() -> CURRENT_TIMESTAMP.

-- ---------------------------------------------------------------------------
-- 1) task_pending_ops  (000041) — durable pending-op queue for wiki/graph tasks
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS task_pending_ops (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id   INTEGER NOT NULL,
    task_type   VARCHAR(64) NOT NULL,
    scope       VARCHAR(32) NOT NULL,
    scope_id    VARCHAR(64) NOT NULL,
    op          VARCHAR(32) NOT NULL,
    dedup_key   VARCHAR(128) NOT NULL DEFAULT '',
    payload     TEXT NOT NULL DEFAULT '{}',
    fail_count  INTEGER NOT NULL DEFAULT 0,
    enqueued_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    claimed_at  DATETIME
);
CREATE INDEX IF NOT EXISTS idx_task_pending_ops_scope
    ON task_pending_ops (task_type, scope, scope_id, id);
CREATE INDEX IF NOT EXISTS idx_task_pending_ops_tenant
    ON task_pending_ops (tenant_id);

-- ---------------------------------------------------------------------------
-- 2) task_dead_letters  (000041) — exhausted-retry task graveyard
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS task_dead_letters (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id   INTEGER NOT NULL,
    task_type   VARCHAR(64) NOT NULL,
    scope       VARCHAR(32) NOT NULL,
    scope_id    VARCHAR(64) NOT NULL,
    related_id  VARCHAR(64) NOT NULL DEFAULT '',
    payload     TEXT NOT NULL,
    last_error  TEXT NOT NULL DEFAULT '',
    fail_count  INTEGER NOT NULL,
    failed_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_task_dead_letters_scope
    ON task_dead_letters (scope, scope_id, failed_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_dead_letters_tenant
    ON task_dead_letters (tenant_id, failed_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_dead_letters_task_type
    ON task_dead_letters (task_type, failed_at DESC);

-- ---------------------------------------------------------------------------
-- 3) wiki_log_entries  (000040) — append-only wiki ingest/retract activity log
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS wiki_log_entries (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id         INTEGER NOT NULL,
    knowledge_base_id VARCHAR(36) NOT NULL,
    action            VARCHAR(32) NOT NULL,
    knowledge_id      VARCHAR(36) NOT NULL DEFAULT '',
    doc_title         TEXT NOT NULL DEFAULT '',
    summary           TEXT NOT NULL DEFAULT '',
    pages_affected    TEXT NOT NULL DEFAULT '[]',
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_wiki_log_entries_kb_id_desc
    ON wiki_log_entries (knowledge_base_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_wiki_log_entries_tenant_id
    ON wiki_log_entries (tenant_id);

-- ---------------------------------------------------------------------------
-- 4) knowledge_processing_spans  (000055) — per-attempt processing trace
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS knowledge_processing_spans (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    knowledge_id    VARCHAR(64) NOT NULL,
    attempt         INTEGER NOT NULL DEFAULT 1,
    span_id         VARCHAR(64) NOT NULL,
    parent_span_id  VARCHAR(64),
    name            VARCHAR(64) NOT NULL,
    kind            VARCHAR(16) NOT NULL,
    status          VARCHAR(16) NOT NULL,
    input           TEXT,
    output          TEXT,
    metadata        TEXT,
    error_code      VARCHAR(64),
    error_message   TEXT,
    error_detail    TEXT,
    started_at      DATETIME,
    finished_at     DATETIME,
    duration_ms     INTEGER,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_kpspan_attempt_span UNIQUE (knowledge_id, attempt, span_id)
);
CREATE INDEX IF NOT EXISTS idx_kpspan_knowledge_attempt
    ON knowledge_processing_spans (knowledge_id, attempt);
CREATE INDEX IF NOT EXISTS idx_kpspan_status_started
    ON knowledge_processing_spans (status, started_at);
CREATE INDEX IF NOT EXISTS idx_kpspan_parent
    ON knowledge_processing_spans (parent_span_id);

-- ---------------------------------------------------------------------------
-- 5) system_settings  (000053) — runtime settings key/value store
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS system_settings (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    key              VARCHAR(128) NOT NULL UNIQUE,
    value            TEXT NOT NULL,
    value_type       VARCHAR(16) NOT NULL,
    category         VARCHAR(32) NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    is_secret        BOOLEAN NOT NULL DEFAULT 0,
    requires_restart BOOLEAN NOT NULL DEFAULT 0,
    last_modified_by VARCHAR(36) NOT NULL DEFAULT '',
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_system_settings_category
    ON system_settings (category);

-- ---------------------------------------------------------------------------
-- 6) organization_members  (000012) — org membership roles
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS organization_members (
    id              VARCHAR(36) PRIMARY KEY,
    organization_id VARCHAR(36) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         VARCHAR(36) NOT NULL,
    tenant_id       INTEGER NOT NULL,
    role            VARCHAR(32) NOT NULL DEFAULT 'viewer',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_org_members_org_user
    ON organization_members(organization_id, user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON organization_members(user_id);
CREATE INDEX IF NOT EXISTS idx_org_members_tenant_id ON organization_members(tenant_id);
CREATE INDEX IF NOT EXISTS idx_org_members_role ON organization_members(role);
