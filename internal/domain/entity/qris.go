// Package entity defines core business entities used across all layers.
package entity

// QrisResult represents the result of QRIS generation.
type QrisResult struct {
	QrisString string
	ImageData  []byte
	Amount     int
	Deskripsi  string
	MessageID  string // Set after sending
}
