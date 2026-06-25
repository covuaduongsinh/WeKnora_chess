-- Down migration for 000002_lite_missing_tables (SQLite / Lite)
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS knowledge_processing_spans;
DROP TABLE IF EXISTS wiki_log_entries;
DROP TABLE IF EXISTS task_dead_letters;
DROP TABLE IF EXISTS task_pending_ops;
