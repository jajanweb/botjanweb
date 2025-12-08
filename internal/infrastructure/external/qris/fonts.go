// Package qris provides QRIS generation and beautiful QR code image rendering.
package qris

import (
	"path/filepath"

	"github.com/fogleman/gg"
)

// Font configuration for the renderer.
var (
	// Primary fonts (Montserrat)
	montserratBold    = "Montserrat-SemiBold.ttf"
	montserratRegular = "Montserrat-Regular.ttf"

	// Fallback fonts (Droid Sans)
	fallbackFontPaths = []string{
		"/usr/share/fonts/google-droid-sans-fonts/DroidSans.ttf",
		"/usr/share/fonts/truetype/droid/DroidSans.ttf",
	}
	fallbackBoldFontPaths = []string{
		"/usr/share/fonts/google-droid-sans-fonts/DroidSans-Bold.ttf",
		"/usr/share/fonts/truetype/droid/DroidSans-Bold.ttf",
	}
)

// loadFont loads Montserrat font with Droid Sans as fallback.
func loadFont(dc *gg.Context, size float64, bold bool, assetsPath string) {
	fontPaths := buildFontPaths(assetsPath, bold)

	for _, path := range fontPaths {
		if err := dc.LoadFontFace(path, size); err == nil {
			return
		}
	}
}

// buildFontPaths builds the list of font paths to try in order.
func buildFontPaths(assetsPath string, bold bool) []string {
	if bold {
		return append(
			[]string{filepath.Join(assetsPath, "fonts", montserratBold)},
			fallbackBoldFontPaths...,
		)
	}
	return append(
		[]string{filepath.Join(assetsPath, "fonts", montserratRegular)},
		fallbackFontPaths...,
	)
}
