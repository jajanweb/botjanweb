// Package entity defines core business entities used across all layers.
package entity

import (
	"fmt"
	"strings"
)

// Product represents available product types.
type Product string

// Product constants - these are the only valid product values.
// Each product maps to a specific sheet in the spreadsheet.
const (
	ProductChatGPT    Product = "ChatGPT"
	ProductGemini     Product = "Gemini"
	ProductYouTube    Product = "YouTube"
	ProductPerplexity Product = "Perplexity"
)

// ProductInfo contains display information for a product.
type ProductInfo struct {
	Key       Product // Short key for input
	FullName  string  // Full display name
	SheetName string  // Target spreadsheet sheet name
}

// Products maps product keys to their full information.
var Products = map[Product]ProductInfo{
	ProductChatGPT: {
		Key:       ProductChatGPT,
		FullName:  "ChatGPT Pro",
		SheetName: "ChatGPT",
	},
	ProductGemini: {
		Key:       ProductGemini,
		FullName:  "GDrive 2TB + Gemini AI Pro",
		SheetName: "Gemini",
	},
	ProductYouTube: {
		Key:       ProductYouTube,
		FullName:  "YouTube Premium",
		SheetName: "YouTube",
	},
	ProductPerplexity: {
		Key:       ProductPerplexity,
		FullName:  "Perplexity Pro",
		SheetName: "Perplexity",
	},
}

// AllProducts returns all valid product keys in order.
func AllProducts() []Product {
	return []Product{
		ProductChatGPT,
		ProductGemini,
		ProductYouTube,
		ProductPerplexity,
	}
}

// ParseProduct parses a string to a valid Product.
// Accepts both short key (e.g., "Gemini") and full name (e.g., "GDrive 2TB + Gemini AI Pro").
// Case-insensitive matching.
func ParseProduct(s string) (Product, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("produk tidak boleh kosong. Pilihan: ChatGPT, Gemini, YouTube, Perplexity")
	}

	lower := strings.ToLower(s)

	// Try exact match with key (case-insensitive)
	for _, p := range AllProducts() {
		if strings.ToLower(string(p)) == lower {
			return p, nil
		}
	}

	// Try partial match with full name
	for _, p := range AllProducts() {
		info := Products[p]
		if strings.ToLower(info.FullName) == lower {
			return p, nil
		}
		// Also try contains for flexibility
		if strings.Contains(strings.ToLower(info.FullName), lower) {
			return p, nil
		}
	}

	// Build error message with valid options
	validOptions := make([]string, 0, len(Products))
	for _, p := range AllProducts() {
		validOptions = append(validOptions, string(p))
	}

	return "", fmt.Errorf("produk '%s' tidak valid. Pilihan: %s", s, strings.Join(validOptions, ", "))
}

// Info returns the ProductInfo for this product.
func (p Product) Info() ProductInfo {
	if info, ok := Products[p]; ok {
		return info
	}
	return ProductInfo{}
}

// FullName returns the full display name.
func (p Product) FullName() string {
	return p.Info().FullName
}

// SheetName returns the target spreadsheet sheet name.
func (p Product) SheetName() string {
	return p.Info().SheetName
}

// String returns the product key as string.
func (p Product) String() string {
	return string(p)
}

// IsValid checks if the product is a valid known product.
func (p Product) IsValid() bool {
	_, ok := Products[p]
	return ok
}
