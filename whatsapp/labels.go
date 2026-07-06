package whatsapp

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"
)

// CreateOrEditLabel creates or edits a WhatsApp label via app state sync and
// upserts the change locally so reads are immediately consistent without
// waiting for the echo event.
func (c *Client) CreateOrEditLabel(ctx context.Context, labelID, name string, color int, deleted bool) error {
	if !c.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}
	patch := appstate.BuildLabelEdit(labelID, name, int32(color), deleted)
	if err := c.wa.SendAppState(ctx, patch); err != nil {
		return fmt.Errorf("failed to send label edit app state: %w", err)
	}
	if c.labelStore != nil {
		if err := c.labelStore.UpsertLabel(labelID, name, color, deleted); err != nil {
			c.log.Warnf("Failed to upsert label %s locally after SendAppState: %v", labelID, err)
		}
	}
	return nil
}

// SetChatLabel applies or removes a label on a chat via app state sync and
// updates the local association store immediately.
func (c *Client) SetChatLabel(ctx context.Context, chatJIDStr, labelID string, labeled bool) error {
	if !c.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}
	targetJID, err := types.ParseJID(chatJIDStr)
	if err != nil {
		return fmt.Errorf("invalid chat JID %q: %w", chatJIDStr, err)
	}
	patch := appstate.BuildLabelChat(targetJID, labelID, labeled)
	if err := c.wa.SendAppState(ctx, patch); err != nil {
		return fmt.Errorf("failed to send label chat app state: %w", err)
	}
	if c.labelStore != nil {
		chatJID := c.normalizeJID(targetJID)
		if err := c.labelStore.UpsertChatAssociation(labelID, chatJID, labeled); err != nil {
			c.log.Warnf("Failed to upsert chat association label=%s chat=%s: %v", labelID, chatJID, err)
		}
	}
	return nil
}
