// Package whatsapp wraps WhatsMeow for WhatsApp connectivity.
package whatsapp

import (
	"context"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// SendText sends a text message to the default group.
func (c *Client) SendText(ctx context.Context, text string) error {
	return c.SendTextTo(ctx, c.groupJID.String(), text)
}

// SendTextTo sends a text message to a specific chat.
func (c *Client) SendTextTo(ctx context.Context, chatID string, text string) error {
	jid, err := types.ParseJID(chatID)
	if err != nil {
		return err
	}

	_, err = c.wm.SendMessage(ctx, jid, &waE2E.Message{
		Conversation: proto.String(text),
	})
	return err
}

// SendTextReply sends a reply to a specific message.
func (c *Client) SendTextReply(ctx context.Context, chatID, text, quotedMsgID, quotedSenderID string) error {
	jid, err := types.ParseJID(chatID)
	if err != nil {
		return err
	}

	_, err = c.wm.SendMessage(ctx, jid, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(quotedMsgID),
				Participant: proto.String(quotedSenderID),
			},
		},
	})
	return err
}

// SendImage sends an image with caption to the default group.
func (c *Client) SendImage(ctx context.Context, imageData []byte, caption string) (string, error) {
	return c.SendImageTo(ctx, c.groupJID.String(), imageData, caption)
}

// SendImageTo sends an image to a specific chat.
func (c *Client) SendImageTo(ctx context.Context, chatID string, imageData []byte, caption string) (string, error) {
	jid, err := types.ParseJID(chatID)
	if err != nil {
		return "", err
	}

	uploaded, err := c.wm.Upload(ctx, imageData, c.getMediaImage())
	if err != nil {
		return "", err
	}

	imgMsg := &waE2E.ImageMessage{
		Caption:       proto.String(caption),
		Mimetype:      proto.String("image/png"),
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(imageData))),
	}

	resp, err := c.wm.SendMessage(ctx, jid, &waE2E.Message{
		ImageMessage: imgMsg,
	})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// SendTextToGroup sends a text message to the configured group, returns message ID.
func (c *Client) SendTextToGroup(ctx context.Context, text string) (string, error) {
	resp, err := c.wm.SendMessage(ctx, c.groupJID, &waE2E.Message{
		Conversation: proto.String(text),
	})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

// SendTextReplyToGroup sends a reply to a specific message in the configured group.
func (c *Client) SendTextReplyToGroup(ctx context.Context, text, quotedMsgID string) error {
	_, err := c.wm.SendMessage(ctx, c.groupJID, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(quotedMsgID),
				Participant: proto.String(c.GetOwnID()),
			},
		},
	})
	return err
}

// SendImageToGroup sends an image to the configured group.
func (c *Client) SendImageToGroup(ctx context.Context, imageData []byte, caption string) (string, error) {
	return c.SendImageTo(ctx, c.groupJID.String(), imageData, caption)
}

// RevokeMessage revokes (deletes for everyone) a message sent by the bot.
// Uses WhatsMeow's BuildRevoke to create a revoke protocol message.
func (c *Client) RevokeMessage(ctx context.Context, chatID, messageID string) error {
	jid, err := types.ParseJID(chatID)
	if err != nil {
		return err
	}

	revokeMsg := c.wm.BuildRevoke(jid, types.EmptyJID, messageID)
	_, err = c.wm.SendMessage(ctx, jid, revokeMsg)
	return err
}
