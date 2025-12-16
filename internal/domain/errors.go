// Package domain contains domain-level errors.
package domain

import "errors"

// Business logic errors.
var (
	ErrQrisGeneration      = errors.New("failed to generate QRIS")
	ErrInvalidNotification = errors.New("invalid notification format")
)

// Family validation errors (Gemini).
var (
	ErrFamilyNotFound = errors.New("family not found in Google Accounts")
	ErrFamilyFull     = errors.New("family is full (5/5 slots used)")
)

// Workspace validation errors (ChatGPT).
var (
	ErrWorkspaceNotFound = errors.New("workspace not found in ChatGPT Accounts")
	ErrWorkspaceFull     = errors.New("workspace is full (4/4 slots used)")
)

// Account errors.
var (
	ErrInvalidAccType = errors.New("invalid account type (choose: Google or ChatGPT)")
)
