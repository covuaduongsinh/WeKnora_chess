-- Down migration for 000003_messages_attachments (SQLite / Lite)
ALTER TABLE messages DROP COLUMN attachments;
