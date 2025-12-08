// Package bot provides WhatsApp bot message parsing and handling.
package parser

import (
	"fmt"
	"strings"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// ParseSelfQrisCommand parses a #qris command for self-QRIS (bot to customer).
// This is used when admin sends QRIS directly to a customer via private chat.
//
// Supports multiline format with required fields:
//
//	#qris 50000
//	produk: gemini
//	nama: Budi Santoso
//	email: budi@email.com
//	family: Personal
//	untuk: 08123456789
//
// Required fields: amount (from first line), produk
// Optional fields: nama, email, family, workspace, paket, kanal, untuk/ke/target
func ParseSelfQrisCommand(text string) (*entity.QrisCommand, error) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(strings.ToLower(text), "#qris") {
		return nil, fmt.Errorf("bukan command #qris")
	}

	rest := strings.TrimSpace(text[5:])
	if rest == "" {
		return nil, fmt.Errorf("nominal wajib diisi")
	}

	lines := strings.Split(rest, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// Parse amount from first line (required)
	amountParts := strings.SplitN(firstLine, " ", 2)
	amount, err := ParseRupiah(amountParts[0])
	if err != nil {
		return nil, fmt.Errorf("nominal tidak valid: %s", amountParts[0])
	}

	cmd := &entity.QrisCommand{
		Amount:     amount,
		IsFormMode: len(lines) > 1, // Has additional form data
		Kanal:      "WhatsApp",     // Default for self-QRIS
	}

	// If first line has description after amount (legacy: #qris 50000 deskripsi)
	if len(amountParts) > 1 {
		cmd.Deskripsi = strings.TrimSpace(amountParts[1])
	}

	// Parse optional form fields from subsequent lines
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
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
			product, err := entity.ParseProduct(value)
			if err == nil {
				cmd.Produk = string(product)
			}
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
		case fieldKanal:
			if value != "" {
				cmd.Kanal = value
			}
		case fieldAkun:
			if value != "" {
				cmd.Akun = value
			}
		case fieldUntuk, "ke", "target", "phone", "nomor":
			// Normalize phone number: remove spaces, dashes, leading +
			value = strings.ReplaceAll(value, " ", "")
			value = strings.ReplaceAll(value, "-", "")
			value = strings.TrimPrefix(value, "+")
			// Convert 08xxx to 628xxx
			if strings.HasPrefix(value, "0") {
				value = "62" + value[1:]
			}
			cmd.TargetPhone = value
		case "deskripsi", "desc", "catatan":
			if value != "" {
				cmd.Deskripsi = value
			}
		}
	}

	// Build description from available data if not explicitly set
	if cmd.Deskripsi == "" && cmd.Nama != "" {
		parts := []string{}
		if cmd.Produk != "" {
			parts = append(parts, cmd.Produk)
		}
		parts = append(parts, cmd.Nama)
		if cmd.Family != "" {
			parts = append(parts, "("+cmd.Family+")")
		} else if cmd.Workspace != "" {
			parts = append(parts, "("+cmd.Workspace+")")
		}
		cmd.Deskripsi = strings.Join(parts, " - ")
	}

	// Validate required field: Produk
	if cmd.Produk == "" {
		return nil, fmt.Errorf("field produk wajib diisi")
	}

	return cmd, nil
}
