// Package domain contains domain-level errors.
package domain

import "errors"

// Command parsing errors.
var (
	ErrInvalidAmount     = errors.New("amount must be a positive number")
	ErrInvalidPartsCount = errors.New("invalid format: must have 5 parts separated by |")
	ErrNotQrisCommand    = errors.New("not a #qris command")
	ErrEmptyField        = errors.New("field cannot be empty")
)

// Business logic errors.
var (
	ErrUnauthorized    = errors.New("user not authorized")
	ErrPaymentNotFound = errors.New("pending payment not found")
	ErrQrisGeneration  = errors.New("failed to generate QRIS")
	ErrSendMessage     = errors.New("failed to send message")
	ErrSheetAppend     = errors.New("failed to log to spreadsheet")
)

// Family validation errors.
var (
	ErrFamilyNotFound = errors.New("family not found in Google Accounts")
	ErrFamilyFull     = errors.New("family is full (5/5 slots used)")
)

// Account errors.
var (
	ErrAccountExists  = errors.New("email already registered")
	ErrInvalidAccType = errors.New("invalid account type (choose: Google or ChatGPT)")
	ErrNotAddAkunCmd  = errors.New("not a #addakun command")
)

// Notification validation errors.
var (
	ErrNotDANAPayment      = errors.New("not a DANA payment notification")
	ErrInvalidNotification = errors.New("invalid notification format")
)
