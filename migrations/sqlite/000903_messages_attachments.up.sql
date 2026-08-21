-- Migration: 000003_messages_attachments (SQLite / Lite)
-- The messages table is missing the `attachments` column (Postgres added it via a
-- versioned ALTER that was never backported to the consolidated SQLite init schema).
-- Without it, saving a chat message fails ("table messages has no column named
-- attachments") and the chat stream returns HTTP 500. JSONB -> TEXT for SQLite.
ALTER TABLE messages ADD COLUMN attachments TEXT DEFAULT '[]';
