// Package whatsapp wraps WhatsMeow for WhatsApp connectivity.
package whatsapp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Connect starts the WhatsApp connection.
// For already logged-in devices, this connects directly.
// For new devices, use the web pairing flow instead of this method.
func (c *Client) Connect(ctx context.Context) error {
	// Add event handler
	c.wm.AddEventHandler(c.eventHandler)

	// Check if already logged in
	if c.wm.Store.ID != nil {
		c.logger.Printf("Logged in as: %s", c.wm.Store.ID.String())
		if err := c.wm.Connect(); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		return nil
	}

	// Not logged in - should use web pairing instead
	c.logger.Println("⚠️ Device not paired. Please use web pairing interface.")
	return nil // Don't error out, let web pairing handle it
}

// Disconnect disconnects from WhatsApp.
func (c *Client) Disconnect() {
	c.wm.Disconnect()
	c.logger.Println("Disconnected from WhatsApp")
}

// Run starts the client and blocks until interrupted.
// If device is not logged in, this will just wait for pairing to complete via web interface.
func (c *Client) Run(ctx context.Context) error {
	// Connect if already logged in
	if c.IsLoggedIn() {
		if err := c.Connect(ctx); err != nil {
			return err
		}
		c.logger.Println("✅ WhatsApp connected and ready")
	} else {
		// Device not logged in - wait for pairing via web interface
		// The pairing flow will call ConnectWithEventHandler() which triggers the connection
		c.logger.Println("⏳ Waiting for device pairing via web interface...")

		// Poll until logged in or context cancelled
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Check if pairing completed
				if c.IsLoggedIn() {
					c.logger.Printf("✅ Device paired successfully: %s", c.GetOwnID())
					c.logger.Println("✅ WhatsApp connected and ready")
					goto RUNNING
				}
			case <-ctx.Done():
				c.logger.Println("Context cancelled during pairing wait")
				return ctx.Err()
			}
		}
	}

RUNNING:
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	c.logger.Println("Bot is running. Press Ctrl+C to stop.")

	select {
	case <-sigChan:
		c.logger.Println("Received interrupt signal, shutting down...")
	case <-ctx.Done():
		c.logger.Println("Context cancelled, shutting down...")
	}

	if c.wm.IsConnected() {
		c.Disconnect()
	}
	return nil
}
