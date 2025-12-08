// Package bot provides WhatsApp bot message parsing and handling.
package parser

import (
	"fmt"
	"strings"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// ParseAddAkunCommand parses a #addakun command string.
// Supports:
// - #addakun → Show general help
// - #addakun google → Show Google account form template
// - #addakun chatgpt → Show ChatGPT account form template
// - #addakun google\n<form data> → Add Google account
// - #addakun chatgpt\n<form data> → Add ChatGPT account
func ParseAddAkunCommand(text string) (*entity.AddAkunCommand, error) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(strings.ToLower(text), "#addakun") {
		return nil, fmt.Errorf("not a #addakun command")
	}

	rest := strings.TrimSpace(text[8:])
	if rest == "" {
		return &entity.AddAkunCommand{IsHelpMode: true}, nil
	}

	// Check for account type parameter (first word)
	accountType := ProductParamUnknown
	lines := strings.SplitN(rest, "\n", 2)
	firstLine := strings.ToLower(strings.TrimSpace(lines[0]))

	switch firstLine {
	case "google":
		accountType = ProductParamGoogle
		if len(lines) == 1 || strings.TrimSpace(lines[1]) == "" || !isFormFormat(lines[1]) {
			return &entity.AddAkunCommand{IsHelpMode: true, AccountType: string(accountType)}, nil
		}
		rest = lines[1]
	case "chatgpt":
		accountType = ProductParamChatGPT
		if len(lines) == 1 || strings.TrimSpace(lines[1]) == "" || !isFormFormat(lines[1]) {
			return &entity.AddAkunCommand{IsHelpMode: true, AccountType: string(accountType)}, nil
		}
		rest = lines[1]
	}

	return parseAddAkunFormFormat(rest, accountType)
}

// parseAddAkunFormFormat parses the form-based format for #addakun.
func parseAddAkunFormFormat(text string, accountType ProductParam) (*entity.AddAkunCommand, error) {
	cmd := &entity.AddAkunCommand{
		AccountType: string(accountType),
	}

	// Set Tipe based on parameter
	switch accountType {
	case ProductParamGoogle:
		cmd.Tipe = entity.AccountTypeGoogle
	case ProductParamChatGPT:
		cmd.Tipe = entity.AccountTypeChatGPT
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
		case fieldTipe:
			accType, ok := entity.ParseAccountType(value)
			if !ok {
				return nil, fmt.Errorf("tipe akun tidak valid: '%s'. Pilihan: Google atau ChatGPT", value)
			}
			cmd.Tipe = accType
		case fieldEmail:
			cmd.Email = value
		case fieldSandi:
			cmd.Sandi = value
		case fieldWorkspace:
			cmd.Workspace = value
		}
	}

	// Validate required fields
	if cmd.Tipe == "" {
		return nil, fmt.Errorf("field 'Tipe' wajib diisi. Pilihan: Google atau ChatGPT")
	}
	if cmd.Email == "" {
		return nil, fmt.Errorf("field 'Email' wajib diisi")
	}
	if cmd.Sandi == "" {
		return nil, fmt.Errorf("field 'Sandi' wajib diisi")
	}

	// Workspace is required for ChatGPT
	if cmd.Tipe == entity.AccountTypeChatGPT && cmd.Workspace == "" {
		return nil, fmt.Errorf("field 'Workspace' wajib diisi untuk akun ChatGPT")
	}

	return cmd, nil
}

// ParseListAkunCommand parses a #listakun command string.
// Supports:
// - #listakun → Show all accounts summary
// - #listakun google → Show only Google accounts
// - #listakun chatgpt → Show only ChatGPT accounts
func ParseListAkunCommand(text string) (*entity.ListAkunCommand, error) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(strings.ToLower(text), "#listakun") {
		return nil, fmt.Errorf("not a #listakun command")
	}

	rest := strings.TrimSpace(text[9:])
	if rest == "" {
		// No filter, show all
		return &entity.ListAkunCommand{}, nil
	}

	// Check for account type filter
	firstLine := strings.ToLower(strings.TrimSpace(rest))

	cmd := &entity.ListAkunCommand{
		AccountType: firstLine,
	}

	switch firstLine {
	case "google":
		cmd.Tipe = entity.AccountTypeGoogle
	case "chatgpt":
		cmd.Tipe = entity.AccountTypeChatGPT
	default:
		// Unknown filter, show help
		cmd.IsHelpMode = true
	}

	return cmd, nil
}
