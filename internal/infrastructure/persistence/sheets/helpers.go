// Package sheets implements Google Sheets repository for data logging.
package sheets

// ptr returns a pointer to the given string.
func ptr(s string) *string {
	return &s
}

// ptr64 returns a pointer to the given float64.
func ptr64(f float64) *float64 {
	return &f
}
