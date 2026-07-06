package storage

import (
	"testing"

	_ "modernc.org/sqlite"
)

func TestUpsertAndListLabels(t *testing.T) {
	db := newTestDB(t)
	store := NewLabelStore(db)

	// insert two labels
	if err := store.UpsertLabel("1", "Work", 0, false); err != nil {
		t.Fatalf("UpsertLabel: %v", err)
	}
	if err := store.UpsertLabel("2", "Personal", 3, false); err != nil {
		t.Fatalf("UpsertLabel: %v", err)
	}
	// insert one deleted label — should NOT appear in ListLabels
	if err := store.UpsertLabel("3", "Old", 1, true); err != nil {
		t.Fatalf("UpsertLabel deleted: %v", err)
	}

	labels, err := store.ListLabels()
	if err != nil {
		t.Fatalf("ListLabels: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("expected 2 non-deleted labels, got %d", len(labels))
	}
	if labels[0].Name != "Work" {
		t.Errorf("got %q, want Work", labels[0].Name)
	}
	if labels[1].Name != "Personal" {
		t.Errorf("got %q, want Personal", labels[1].Name)
	}
}

func TestUpsertLabelUpdatesExisting(t *testing.T) {
	db := newTestDB(t)
	store := NewLabelStore(db)

	if err := store.UpsertLabel("1", "Work", 0, false); err != nil {
		t.Fatalf("UpsertLabel: %v", err)
	}
	// rename it
	if err := store.UpsertLabel("1", "Work Updated", 5, false); err != nil {
		t.Fatalf("UpsertLabel update: %v", err)
	}

	labels, err := store.ListLabels()
	if err != nil {
		t.Fatalf("ListLabels: %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("expected 1, got %d", len(labels))
	}
	if labels[0].Name != "Work Updated" {
		t.Errorf("got %q, want Work Updated", labels[0].Name)
	}
	if labels[0].Color != 5 {
		t.Errorf("color: got %d, want 5", labels[0].Color)
	}
}

func TestChatAssociationAndGetChatLabels(t *testing.T) {
	db := newTestDB(t)
	store := NewLabelStore(db)

	if err := store.UpsertLabel("1", "Work", 0, false); err != nil {
		t.Fatalf("UpsertLabel: %v", err)
	}
	if err := store.UpsertLabel("2", "Personal", 0, false); err != nil {
		t.Fatalf("UpsertLabel: %v", err)
	}

	chat := "5511999@s.whatsapp.net"
	if err := store.UpsertChatAssociation("1", chat, true); err != nil {
		t.Fatalf("UpsertChatAssociation: %v", err)
	}
	if err := store.UpsertChatAssociation("2", chat, true); err != nil {
		t.Fatalf("UpsertChatAssociation 2: %v", err)
	}

	labels, err := store.GetChatLabels(chat)
	if err != nil {
		t.Fatalf("GetChatLabels: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels for chat, got %d", len(labels))
	}

	// un-label
	if err := store.UpsertChatAssociation("1", chat, false); err != nil {
		t.Fatalf("UpsertChatAssociation unlabel: %v", err)
	}
	labels, err = store.GetChatLabels(chat)
	if err != nil {
		t.Fatalf("GetChatLabels after unlabel: %v", err)
	}
	if len(labels) != 1 {
		t.Fatalf("expected 1 label after unlabel, got %d", len(labels))
	}
	if labels[0].LabelID != "2" {
		t.Errorf("expected label 2, got %s", labels[0].LabelID)
	}
}

func TestListChatsByLabel(t *testing.T) {
	db := newTestDB(t)
	store := NewLabelStore(db)

	if err := store.UpsertLabel("1", "Work", 0, false); err != nil {
		t.Fatalf("UpsertLabel: %v", err)
	}

	chats := []string{
		"a@s.whatsapp.net",
		"b@s.whatsapp.net",
		"c@s.whatsapp.net",
	}
	for _, c := range chats {
		if err := store.UpsertChatAssociation("1", c, true); err != nil {
			t.Fatalf("UpsertChatAssociation %s: %v", c, err)
		}
	}

	got, err := store.ListChatsByLabel("1")
	if err != nil {
		t.Fatalf("ListChatsByLabel: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 chats, got %d", len(got))
	}
}

func TestNextLabelID(t *testing.T) {
	db := newTestDB(t)
	store := NewLabelStore(db)

	// empty table → should return "1"
	id, err := store.NextLabelID()
	if err != nil {
		t.Fatalf("NextLabelID (empty): %v", err)
	}
	if id != "1" {
		t.Errorf("expected '1', got %q", id)
	}

	// insert a label with id "5"
	if err := store.UpsertLabel("5", "Five", 0, false); err != nil {
		t.Fatalf("UpsertLabel: %v", err)
	}

	id, err = store.NextLabelID()
	if err != nil {
		t.Fatalf("NextLabelID after insert: %v", err)
	}
	if id != "6" {
		t.Errorf("expected '6', got %q", id)
	}
}

func TestGetLabelByID(t *testing.T) {
	db := newTestDB(t)
	store := NewLabelStore(db)

	// not found
	l, err := store.GetLabelByID("99")
	if err != nil {
		t.Fatalf("GetLabelByID not found: %v", err)
	}
	if l != nil {
		t.Fatal("expected nil for missing label")
	}

	if err := store.UpsertLabel("7", "Lucky", 2, false); err != nil {
		t.Fatalf("UpsertLabel: %v", err)
	}
	l, err = store.GetLabelByID("7")
	if err != nil {
		t.Fatalf("GetLabelByID: %v", err)
	}
	if l == nil {
		t.Fatal("expected label, got nil")
	}
	if l.Name != "Lucky" || l.Color != 2 {
		t.Errorf("unexpected label: %+v", l)
	}
}

func TestMessageAssociation(t *testing.T) {
	db := newTestDB(t)
	store := NewLabelStore(db)

	if err := store.UpsertLabel("1", "Work", 0, false); err != nil {
		t.Fatalf("UpsertLabel: %v", err)
	}

	chat := "chat@s.whatsapp.net"
	msgID := "msg-abc-123"

	if err := store.UpsertMessageAssociation("1", chat, msgID, true); err != nil {
		t.Fatalf("UpsertMessageAssociation: %v", err)
	}

	// message association should NOT affect GetChatLabels (different message_id)
	labels, err := store.GetChatLabels(chat)
	if err != nil {
		t.Fatalf("GetChatLabels: %v", err)
	}
	if len(labels) != 0 {
		t.Errorf("expected 0 chat labels (message-level assoc), got %d", len(labels))
	}

	// un-label the message
	if err := store.UpsertMessageAssociation("1", chat, msgID, false); err != nil {
		t.Fatalf("UpsertMessageAssociation unlabel: %v", err)
	}

	// verify it is stored (labeled=0)
	var labeled int
	err = db.QueryRow(`SELECT labeled FROM label_associations WHERE label_id='1' AND chat_jid=? AND message_id=?`, chat, msgID).Scan(&labeled)
	if err != nil {
		t.Fatalf("readback: %v", err)
	}
	if labeled != 0 {
		t.Errorf("expected labeled=0, got %d", labeled)
	}
}
