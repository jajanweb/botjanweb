// Package whatsapp wraps WhatsMeow for WhatsApp connectivity.
package whatsapp

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow"
)

// IsLoggedIn checks if the client is currently logged in.
func (c *Client) IsLoggedIn() bool {
	return c.wm.Store.ID != nil
}

// GetQRChannelForPairing returns a channel for QR code pairing events.
// This is used by the web pairing interface.
// IMPORTANT: This must be called BEFORE Connect() is called.
func (c *Client) GetQRChannelForPairing(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error) {
	// Check if already connected - this is an error condition
	if c.wm.IsConnected() {
		return nil, fmt.Errorf("cannot get QR channel: websocket is already connected")
	}

	// Get QR channel (this must be called before Connect)
	qrChan, err := c.wm.GetQRChannel(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get QR channel: %w", err)
	}

	return qrChan, nil
}

// ConnectWithEventHandler connects the websocket with event handler already set.
// This is used by the web pairing flow.
func (c *Client) ConnectWithEventHandler() error {
	if c.wm.IsConnected() {
		return nil // Already connected
	}

	// Add event handler before connecting
	c.wm.AddEventHandler(c.eventHandler)

	if err := c.wm.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	return nil
}
