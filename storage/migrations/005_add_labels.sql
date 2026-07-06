-- Migration: 005_add_labels
-- Description: add labels and label associations
-- Previous: 004
-- Version: 005
-- Created: 2026-07-06

-- Labels created or synced from WhatsApp app state.
CREATE TABLE IF NOT EXISTS labels (
    label_id   TEXT    PRIMARY KEY,
    name       TEXT    NOT NULL DEFAULT '',
    color      INTEGER NOT NULL DEFAULT 0,
    deleted    INTEGER NOT NULL DEFAULT 0,
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- Associations between labels and chats/messages.
-- chat_jid = '' is not used here; message_id = '' means a chat-level association.
CREATE TABLE IF NOT EXISTS label_associations (
    label_id   TEXT    NOT NULL,
    chat_jid   TEXT    NOT NULL DEFAULT '',
    message_id TEXT    NOT NULL DEFAULT '',
    labeled    INTEGER NOT NULL DEFAULT 1,
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (label_id, chat_jid, message_id)
);

CREATE INDEX IF NOT EXISTS idx_label_assoc_chat  ON label_associations(chat_jid)  WHERE labeled = 1;
CREATE INDEX IF NOT EXISTS idx_label_assoc_label ON label_associations(label_id)  WHERE labeled = 1;
