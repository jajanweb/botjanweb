// Package parser provides command and form parsing utilities.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// ParseQrisCommand parses a #qris command string for group orders.
// Supports:
// - #qris → Show general help
// - #qris google → Show Gemini form template
// - #qris chatgpt → Show ChatGPT form template
// - #qris google\n<form data> → Process Gemini order
// - #qris chatgpt\n<form data> → Process ChatGPT order
// - Legacy: #qris 25000 desc (self-QRIS only, not for group)
func ParseQrisCommand(text string, defaultKanal string) (*entity.QrisCommand, error) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(strings.ToLower(text), "#qris") {
		return nil, fmt.Errorf("not a #qris command")
	}

	rest := strings.TrimSpace(text[5:])
	if rest == "" {
		return &entity.QrisCommand{IsHelpMode: true}, nil
	}

	// Check for product parameter (first word)
	productType := ProductParamUnknown
	lines := strings.SplitN(rest, "\n", 2)
	firstLine := strings.ToLower(strings.TrimSpace(lines[0]))

	switch firstLine {
	case "google":
		productType = ProductParamGoogle
		if len(lines) == 1 || strings.TrimSpace(lines[1]) == "" || !isFormFormat(lines[1]) {
			return &entity.QrisCommand{IsHelpMode: true, ProductType: string(productType)}, nil
		}
		rest = lines[1]
	case "chatgpt":
		productType = ProductParamChatGPT
		if len(lines) == 1 || strings.TrimSpace(lines[1]) == "" || !isFormFormat(lines[1]) {
			return &entity.QrisCommand{IsHelpMode: true, ProductType: string(productType)}, nil
		}
		rest = lines[1]
	}

	if isFormFormat(rest) {
		return parseQrisFormFormat(rest, productType, defaultKanal)
	}

	// Legacy format: #qris <amount> <description>
	return parseQrisLegacyFormat(rest)
}

// parseQrisFormFormat parses the form-based format for group orders.
func parseQrisFormFormat(text string, productType ProductParam, defaultKanal string) (*entity.QrisCommand, error) {
	cmd := &entity.QrisCommand{
		IsFormMode:  true,
		ProductType: string(productType),
		Kanal:       defaultKanal,
	}

	// Set product based on productType parameter
	switch productType {
	case ProductParamGoogle:
		cmd.Produk = string(entity.ProductGemini)
	case ProductParamChatGPT:
		cmd.Produk = string(entity.ProductChatGPT)
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || isDecorator(line) {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		field := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch field {
		case fieldProduk:
			// Legacy: Allow override via Produk field
			product, err := entity.ParseProduct(value)
			if err != nil {
				return nil, err
			}
			cmd.Produk = string(product)
		case fieldNama:
			cmd.Nama = value
		case fieldEmail:
			cmd.Email = value
		case fieldFamily:
			cmd.Family = value
		case fieldWorkspace:
			cmd.Workspace = value
		case fieldPaket:
			cmd.Paket = value
		case fieldNominal:
			// Remove thousands separator
			amountStr := strings.ReplaceAll(value, ".", "")
			amountStr = strings.ReplaceAll(amountStr, ",", "")
			amount, err := strconv.Atoi(amountStr)
			if err != nil {
				return nil, fmt.Errorf("nominal tidak valid: %s", value)
			}
			if amount <= 0 {
				return nil, fmt.Errorf("nominal harus lebih dari 0")
			}
			cmd.Amount = amount
		case fieldKanal:
			if value != "" {
				cmd.Kanal = value
			}
		case fieldAkun:
			if value != "" {
				cmd.Akun = value
			}
		}
	}

	// Validate required fields
	if cmd.Produk == "" {
		return nil, fmt.Errorf("field 'Produk' wajib diisi. Pilihan: ChatGPT, Gemini, YouTube, Perplexity")
	}
	if cmd.Nama == "" {
		return nil, fmt.Errorf("field 'Nama' wajib diisi")
	}
	if cmd.Email == "" {
		return nil, fmt.Errorf("field 'Email' wajib diisi")
	}
	if cmd.Amount <= 0 {
		return nil, fmt.Errorf("field 'Nominal' wajib diisi dengan nilai positif")
	}

	// Product-specific validation and description building
	switch productType {
	case ProductParamGoogle:
		if cmd.Family == "" {
			return nil, fmt.Errorf("field 'Family' wajib diisi untuk produk Gemini")
		}
		cmd.Deskripsi = fmt.Sprintf("%s - %s (%s)", cmd.Produk, cmd.Nama, cmd.Family)
	case ProductParamChatGPT:
		if cmd.Workspace == "" {
			return nil, fmt.Errorf("field 'Workspace' wajib diisi untuk produk ChatGPT")
		}
		if cmd.Paket == "" {
			return nil, fmt.Errorf("field 'Paket' wajib diisi untuk produk ChatGPT")
		}
		cmd.Deskripsi = fmt.Sprintf("%s - %s (%s)", cmd.Produk, cmd.Nama, cmd.Workspace)
	default:
		// Other products require Family
		if cmd.Family == "" {
			return nil, fmt.Errorf("field 'Family' wajib diisi")
		}
		cmd.Deskripsi = fmt.Sprintf("%s - %s (%s)", cmd.Produk, cmd.Nama, cmd.Family)
	}

	return cmd, nil
}

// parseQrisLegacyFormat parses the old format: #qris <amount> <description>
// Note: Only used for self-QRIS, not allowed in group orders.
func parseQrisLegacyFormat(rest string) (*entity.QrisCommand, error) {
	parts := strings.SplitN(rest, " ", 2)

	amount, err := ParseRupiah(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %s", parts[0])
	}

	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	var deskripsi string
	if len(parts) > 1 {
		deskripsi = strings.TrimSpace(parts[1])
	}

	return &entity.QrisCommand{
		Amount:    amount,
		Deskripsi: deskripsi,
	}, nil
}
