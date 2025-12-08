// Package whatsapp wraps WhatsMeow for WhatsApp connectivity.
package whatsapp

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/exernia/botjanweb/pkg/logger"

	_ "github.com/lib/pq"           // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3" // SQLite driver

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// MessageHandler is called when a new message is received.
type MessageHandler func(ctx context.Context, msg *entity.Message)

// Client wraps WhatsMeow client implementing MessagingPort.
type Client struct {
	wm       *whatsmeow.Client
	groupJID types.JID
	handler  MessageHandler
	logger   *log.Logger
}

// NewClient creates and initializes a new WhatsMeow client.
// Supports both PostgreSQL (via DATABASE_URL) and SQLite (via dbPath).
func NewClient(ctx context.Context, dbPath string, groupJIDStr string) (*Client, error) {
	log := logger.WhatsApp

	groupJID, err := types.ParseJID(groupJIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID '%s': %w", groupJIDStr, err)
	}

	dbLog := waLog.Stdout("Database", "ERROR", true)

	// Determine database type and connection string
	var container *sqlstore.Container
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		// Use PostgreSQL (production/Heroku)
		log.Println("📊 Using PostgreSQL database")
		container, err = sqlstore.New(ctx, "postgres", databaseURL, dbLog)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL container: %w", err)
		}
	} else {
		// Use SQLite (local development)
		log.Printf("📊 Using SQLite database: %s", dbPath)
		container, err = sqlstore.New(ctx, "sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=10000", dbPath), dbLog)
		if err != nil {
			return nil, fmt.Errorf("failed to create SQLite container: %w", err)
		}
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device store: %w", err)
	}

	// Set log level to WARN to suppress SessionCipher errors (old counter messages)
	// These are expected when messages arrive out of order and are handled internally by WhatsMeow
	clientLog := waLog.Stdout("Client", "WARN", true)
	wmClient := whatsmeow.NewClient(deviceStore, clientLog)

	return &Client{
		wm:       wmClient,
		groupJID: groupJID,
		logger:   log,
	}, nil
}

// SetMessageHandler sets the callback for incoming messages.
func (c *Client) SetMessageHandler(handler MessageHandler) {
	c.handler = handler
}

// GetOwnID returns the bot's own ID.
func (c *Client) GetOwnID() string {
	if c.wm.Store.ID == nil {
		return ""
	}
	return c.wm.Store.ID.String()
}

// GetGroupJID returns the configured group JID string.
func (c *Client) GetGroupJID() string {
	return c.groupJID.String()
}

// getMediaImage returns the media type for image uploads.
func (c *Client) getMediaImage() whatsmeow.MediaType {
	return whatsmeow.MediaImage
}
