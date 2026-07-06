package storage

import (
	"database/sql"
	"fmt"
)

// Label represents a WhatsApp label.
type Label struct {
	LabelID string
	Name    string
	Color   int
	Deleted bool
}

// LabelStore handles label and label-association operations on the database.
type LabelStore struct {
	db *sql.DB
}

// NewLabelStore creates a new label store instance.
func NewLabelStore(db *sql.DB) *LabelStore {
	return &LabelStore{db: db}
}

// UpsertLabel inserts or updates a label record.
func (s *LabelStore) UpsertLabel(labelID, name string, color int, deleted bool) error {
	const query = `
	INSERT INTO labels (label_id, name, color, deleted, updated_at)
	VALUES (?, ?, ?, ?, unixepoch())
	ON CONFLICT(label_id) DO UPDATE SET
		name       = excluded.name,
		color      = excluded.color,
		deleted    = excluded.deleted,
		updated_at = excluded.updated_at
	`
	deletedInt := 0
	if deleted {
		deletedInt = 1
	}
	_, err := s.db.Exec(query, labelID, name, color, deletedInt)
	if err != nil {
		return fmt.Errorf("failed to upsert label %s: %w", labelID, err)
	}
	return nil
}

// UpsertChatAssociation inserts or updates a label–chat association.
// message_id is always '' for chat-level associations.
func (s *LabelStore) UpsertChatAssociation(labelID, chatJID string, labeled bool) error {
	const query = `
	INSERT INTO label_associations (label_id, chat_jid, message_id, labeled, updated_at)
	VALUES (?, ?, '', ?, unixepoch())
	ON CONFLICT(label_id, chat_jid, message_id) DO UPDATE SET
		labeled    = excluded.labeled,
		updated_at = excluded.updated_at
	`
	labeledInt := 0
	if labeled {
		labeledInt = 1
	}
	_, err := s.db.Exec(query, labelID, chatJID, labeledInt)
	if err != nil {
		return fmt.Errorf("failed to upsert chat association label=%s chat=%s: %w", labelID, chatJID, err)
	}
	return nil
}

// UpsertMessageAssociation inserts or updates a label–message association.
func (s *LabelStore) UpsertMessageAssociation(labelID, chatJID, messageID string, labeled bool) error {
	const query = `
	INSERT INTO label_associations (label_id, chat_jid, message_id, labeled, updated_at)
	VALUES (?, ?, ?, ?, unixepoch())
	ON CONFLICT(label_id, chat_jid, message_id) DO UPDATE SET
		labeled    = excluded.labeled,
		updated_at = excluded.updated_at
	`
	labeledInt := 0
	if labeled {
		labeledInt = 1
	}
	_, err := s.db.Exec(query, labelID, chatJID, messageID, labeledInt)
	if err != nil {
		return fmt.Errorf("failed to upsert message association label=%s chat=%s msg=%s: %w", labelID, chatJID, messageID, err)
	}
	return nil
}

// ListLabels returns all non-deleted labels ordered by label_id.
func (s *LabelStore) ListLabels() ([]Label, error) {
	const query = `
	SELECT label_id, name, color, deleted
	FROM labels
	WHERE deleted = 0
	ORDER BY label_id
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		var l Label
		var deletedInt int
		if err := rows.Scan(&l.LabelID, &l.Name, &l.Color, &deletedInt); err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		l.Deleted = deletedInt != 0
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// GetChatLabels returns all non-deleted labels currently applied to a chat.
func (s *LabelStore) GetChatLabels(chatJID string) ([]Label, error) {
	const query = `
	SELECT l.label_id, l.name, l.color, l.deleted
	FROM labels l
	JOIN label_associations a ON l.label_id = a.label_id
	WHERE a.chat_jid = ? AND a.message_id = '' AND a.labeled = 1 AND l.deleted = 0
	ORDER BY l.label_id
	`
	rows, err := s.db.Query(query, chatJID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat labels: %w", err)
	}
	defer rows.Close()

	var labels []Label
	for rows.Next() {
		var l Label
		var deletedInt int
		if err := rows.Scan(&l.LabelID, &l.Name, &l.Color, &deletedInt); err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		l.Deleted = deletedInt != 0
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

// ListChatsByLabel returns all chat JIDs that currently carry the given label.
func (s *LabelStore) ListChatsByLabel(labelID string) ([]string, error) {
	const query = `
	SELECT DISTINCT chat_jid
	FROM label_associations
	WHERE label_id = ? AND message_id = '' AND labeled = 1
	ORDER BY chat_jid
	`
	rows, err := s.db.Query(query, labelID)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats by label: %w", err)
	}
	defer rows.Close()

	var jids []string
	for rows.Next() {
		var jid string
		if err := rows.Scan(&jid); err != nil {
			return nil, fmt.Errorf("failed to scan jid: %w", err)
		}
		jids = append(jids, jid)
	}
	return jids, rows.Err()
}

// GetLabelByID returns a single label by its ID, or nil if not found.
func (s *LabelStore) GetLabelByID(labelID string) (*Label, error) {
	const query = `
	SELECT label_id, name, color, deleted
	FROM labels
	WHERE label_id = ?
	`
	var l Label
	var deletedInt int
	err := s.db.QueryRow(query, labelID).Scan(&l.LabelID, &l.Name, &l.Color, &deletedInt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get label %s: %w", labelID, err)
	}
	l.Deleted = deletedInt != 0
	return &l, nil
}

// NextLabelID computes max(CAST(label_id AS INTEGER)) + 1 and returns it as a
// string, starting from "1" when the table is empty. Only numeric label IDs
// are considered; non-numeric IDs are ignored.
func (s *LabelStore) NextLabelID() (string, error) {
	const query = `
	SELECT COALESCE(MAX(CAST(label_id AS INTEGER)), 0) + 1
	FROM labels
	WHERE CAST(label_id AS INTEGER) > 0
	`
	var next int
	if err := s.db.QueryRow(query).Scan(&next); err != nil {
		return "", fmt.Errorf("failed to compute next label id: %w", err)
	}
	return fmt.Sprintf("%d", next), nil
}
