// Package qris provides QRIS generation infrastructure.
package qris

import (
	"fmt"

	goqris "github.com/fyvri/go-qris/pkg/services"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/exernia/botjanweb/internal/domain/entity"
)

// DefaultQRSize is the default QR code image size in pixels.
const DefaultQRSize = 256

// Generator handles QRIS conversion and QR code image generation.
// Implements usecase.QrisGenerator interface.
type Generator struct {
	qrisService goqris.QRISInterface
	renderer    *Renderer
}

// NewGenerator creates a new QRIS generator instance.
// assetsPath is the path to the assets folder containing template.png and logo.png.
func NewGenerator(assetsPath string) *Generator {
	return &Generator{
		qrisService: goqris.NewQRIS(),
		renderer:    NewRenderer(assetsPath),
	}
}

// GenerateQris generates a dynamic QRIS with template image.
// Implements usecase.QrisGenerator interface.
func (g *Generator) GenerateQris(baseQR string, amount int, description string) (*entity.QrisResult, error) {
	if baseQR == "" {
		return nil, fmt.Errorf("base QRIS payload cannot be empty")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive, got %d", amount)
	}

	// Convert static QRIS to dynamic with amount
	// Parameters: qrisString, merchantCity, merchantPostalCode, paymentAmount, paymentFeeCategory, paymentFee, terminalLabel
	dynamicQRIS, err, _ := g.qrisService.Convert(baseQR, "", "", amount, "", 0, "")
	if err != nil {
		return nil, fmt.Errorf("failed to convert QRIS to dynamic: %w", err)
	}

	// Render beautiful QR image with template
	pngBytes, err := g.renderer.RenderQRISImage(RenderParams{
		QRISString: dynamicQRIS,
		Amount:     amount,
		Deskripsi:  description,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render image: %w", err)
	}

	return &entity.QrisResult{
		QrisString: dynamicQRIS,
		ImageData:  pngBytes,
		Amount:     amount,
		Deskripsi:  description,
	}, nil
}

// BuildDynamicQRIS converts a static QRIS payload to a dynamic one with the specified amount.
func (g *Generator) BuildDynamicQRIS(baseQR string, amount int) (string, error) {
	if baseQR == "" {
		return "", fmt.Errorf("base QRIS payload cannot be empty")
	}
	if amount <= 0 {
		return "", fmt.Errorf("amount must be positive, got %d", amount)
	}

	dynamicQRIS, err, _ := g.qrisService.Convert(baseQR, "", "", amount, "", 0, "")
	if err != nil {
		return "", fmt.Errorf("failed to convert QRIS to dynamic: %w", err)
	}

	return dynamicQRIS, nil
}

// BuildQRISPNG generates a plain QR code PNG image from a QRIS string.
func (g *Generator) BuildQRISPNG(qrisString string) ([]byte, error) {
	return g.BuildQRISPNGWithSize(qrisString, DefaultQRSize)
}

// BuildQRISPNGWithSize generates a plain QR code PNG image with a custom size.
func (g *Generator) BuildQRISPNGWithSize(qrisString string, size int) ([]byte, error) {
	if qrisString == "" {
		return nil, fmt.Errorf("QRIS string cannot be empty")
	}
	if size <= 0 {
		size = DefaultQRSize
	}

	png, err := qrcode.Encode(qrisString, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return png, nil
}
