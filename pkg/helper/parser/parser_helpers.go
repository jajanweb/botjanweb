// Package parser provides command and form parsing utilities.
package parser

import (
	"strings"
)

// isFormFormat checks if text contains form format (has field: value pattern).
func isFormFormat(text string) bool {
	formFields := map[string]bool{
		fieldProduk: true, fieldNama: true, fieldEmail: true,
		fieldFamily: true, fieldNominal: true, fieldKanal: true,
		fieldAkun: true, fieldWorkspace: true, fieldPaket: true,
		fieldTipe: true, fieldSandi: true, fieldUntuk: true,
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || isDecorator(line) {
			continue
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			field := strings.ToLower(strings.TrimSpace(line[:idx]))
			if formFields[field] {
				return true
			}
		}
	}
	return false
}

// isDecorator checks if a line is a visual decorator (dashes, equals, etc.).
func isDecorator(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return true
	}
	decorators := "─-=━═┄┈┅┉"
	nonDecoratorCount := 0
	for _, r := range line {
		if !strings.ContainsRune(decorators, r) {
			nonDecoratorCount++
		}
	}
	return nonDecoratorCount < len(line)/5+1
}
