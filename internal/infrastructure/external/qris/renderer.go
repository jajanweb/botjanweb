// Package qris provides QRIS generation and beautiful QR code image rendering.
package qris

import (
	"bytes"
	"fmt"
	"image/png"
	"path/filepath"

	"github.com/fogleman/gg"
)

// Renderer configuration constants.
const (
	// Template dimensions
	TemplateWidth  = 600
	TemplateHeight = 900

	// QR Code positioning and size
	QRCenterX    = 300
	QRCenterY    = 390
	QRSize       = 452
	QRPadding    = 2
	BorderRadius = 24.0

	// Text positioning
	AmountCenterX     = 300
	AmountCenterY     = 702
	AmountFontSize    = 40.0
	DeskripsiFontSize = 24.0
	DeskripsiCenterX  = 300
	DeskripsiCenterY  = 804

	// Description constraints
	DeskripsiMaxWidth = 472 // Maximum width in pixels
	DeskripsiMaxChars = 96  // Maximum characters before truncation
	DeskripsiMaxLines = 3   // Maximum lines to display

	// Text color - #061B30
	TextColorR = 6
	TextColorG = 27
	TextColorB = 48
)

// Renderer handles beautiful QRIS image generation using template overlay.
type Renderer struct {
	assetsPath string
}

// NewRenderer creates a new QRIS image renderer.
// assetsPath should be the path to the assets folder containing template.png and logo.png.
func NewRenderer(assetsPath string) *Renderer {
	return &Renderer{
		assetsPath: assetsPath,
	}
}

// RenderParams contains parameters for rendering a QRIS image.
type RenderParams struct {
	QRISString string // The QRIS payload to encode
	Amount     int    // Transaction amount in IDR
	Deskripsi  string // Transaction description (optional)
}

// RenderQRISImage generates a beautifully designed QRIS payment image.
func (r *Renderer) RenderQRISImage(params RenderParams) ([]byte, error) {
	// Load template image
	templateImg, err := loadImage(r.getTemplatePath())
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Create drawing context from template
	dc := gg.NewContextForImage(templateImg)

	// Generate and draw QR code with logo
	if err := r.drawQRCode(dc, params.QRISString); err != nil {
		return nil, fmt.Errorf("failed to draw QR code: %w", err)
	}

	// Draw amount text
	r.drawAmount(dc, params.Amount)

	// Draw description if present
	if params.Deskripsi != "" {
		r.drawDeskripsi(dc, params.Deskripsi)
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}

// getTemplatePath returns the path to the template image.
func (r *Renderer) getTemplatePath() string {
	return filepath.Join(r.assetsPath, "template.png")
}

// getLogoPath returns the path to the logo image.
func (r *Renderer) getLogoPath() string {
	return filepath.Join(r.assetsPath, "logo.png")
}
