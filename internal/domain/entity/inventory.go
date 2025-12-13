// Package entity defines core business entities used across all layers.
package entity

// SlotInfo represents availability information for a family/workspace.
type SlotInfo struct {
	Name          string // Family name or Workspace name
	Product       string // "ChatGPT" or "Gemini"
	TotalSlots    int    // Maximum slots allowed
	UsedSlots     int    // Currently filled slots
	AvailableSlot int    // Remaining available slots
}

// SlotAvailabilityResult contains the result of checking slot availability.
type SlotAvailabilityResult struct {
	Product       string     // "ChatGPT" or "Gemini"
	Slots         []SlotInfo // List of families/workspaces with availability
	TotalEntries  int        // Total families/workspaces found
	AvailableOnly bool       // Whether to show only available ones
}

// RedeemCodeInfo represents a Perplexity redeem code entry.
type RedeemCodeInfo struct {
	No              int    // Row number
	Email           string // Email account
	KodeRedeem      string // Redeem code
	TanggalAktivasi string // Activation date (empty = available)
	TanggalBerakhir string // Expiry date
}

// RedeemCodeResult contains the result of checking redeem code availability.
type RedeemCodeResult struct {
	Codes          []RedeemCodeInfo // List of codes (all or available only)
	TotalCodes     int              // Total codes in sheet
	AvailableCodes int              // Codes not yet activated
	AvailableOnly  bool             // Whether showing only available ones
}
