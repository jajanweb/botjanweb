// Package qris provides QRIS generation and beautiful QR code image rendering.
package qris

import (
	"image"
	"image/color"
	"os"

	"github.com/exernia/botjanweb/pkg/helper/formatter"
	"github.com/fogleman/gg"
	qrcode "github.com/skip2/go-qrcode"
)

// drawQRCode generates and draws the QR code with rounded background and center logo.
func (r *Renderer) drawQRCode(dc *gg.Context, qrisString string) error {
	// Generate QR code
	qr, err := qrcode.New(qrisString, qrcode.Medium)
	if err != nil {
		return err
	}

	// Calculate positions
	qrX := float64(QRCenterX - QRSize/2)
	qrY := float64(QRCenterY - QRSize/2)
	qrSize := float64(QRSize)

	// Draw white rounded rectangle background
	dc.SetColor(color.White)
	drawRoundedRect(dc, qrX, qrY, qrSize, qrSize, BorderRadius)
	dc.Fill()

	// Generate QR image (slightly smaller for padding)
	innerQRSize := QRSize - (QRPadding * 2)
	qrImage := qr.Image(innerQRSize)

	// Draw QR code centered
	qrDrawX := QRCenterX - innerQRSize/2
	qrDrawY := QRCenterY - innerQRSize/2
	dc.DrawImage(qrImage, qrDrawX, qrDrawY)

	// Draw logo in center of QR (ignore error - QR works without logo)
	_ = r.drawLogo(dc)

	// Draw subtle border
	dc.SetColor(color.RGBA{220, 220, 220, 255})
	dc.SetLineWidth(1)
	drawRoundedRect(dc, qrX, qrY, qrSize, qrSize, BorderRadius)
	dc.Stroke()

	return nil
}

// drawLogo draws the logo in the center of the QR code.
func (r *Renderer) drawLogo(dc *gg.Context) error {
	logoImg, err := loadImage(r.getLogoPath())
	if err != nil {
		return err
	}

	// White circle background for logo
	logoSize := float64(QRSize) * 0.19
	dc.SetColor(color.White)
	dc.DrawCircle(float64(QRCenterX), float64(QRCenterY), logoSize/2+6)
	dc.Fill()

	// Draw logo centered
	dc.DrawImageAnchored(logoImg, QRCenterX, QRCenterY, 0.5, 0.5)

	return nil
}

// drawAmount draws the transaction amount.
func (r *Renderer) drawAmount(dc *gg.Context, amount int) {
	formattedAmount := formatter.FormatRupiah(amount)
	loadFont(dc, AmountFontSize, true, r.assetsPath)
	dc.SetColor(color.RGBA{TextColorR, TextColorG, TextColorB, 255})
	dc.DrawStringAnchored(formattedAmount, float64(AmountCenterX), float64(AmountCenterY), 0.5, 0.5)
}

// drawDeskripsi draws the transaction description with word wrap and truncation.
func (r *Renderer) drawDeskripsi(dc *gg.Context, deskripsi string) {
	loadFont(dc, DeskripsiFontSize, false, r.assetsPath)
	dc.SetColor(color.RGBA{TextColorR, TextColorG, TextColorB, 255})

	// Truncate if exceeds max characters
	text := formatter.TruncateWithEllipsis(deskripsi, DeskripsiMaxChars)

	// Word wrap within max width
	lines := dc.WordWrap(text, DeskripsiMaxWidth)

	// Limit to max lines, add ellipsis if truncated
	if len(lines) > DeskripsiMaxLines {
		lines = lines[:DeskripsiMaxLines]
		// Add ellipsis to last line if we had to truncate lines
		lastLine := lines[DeskripsiMaxLines-1]
		if len(lastLine) > 3 {
			lines[DeskripsiMaxLines-1] = lastLine[:len(lastLine)-3] + "..."
		}
	}

	// Draw lines centered
	lineHeight := DeskripsiFontSize * 1.3
	for i, line := range lines {
		y := float64(DeskripsiCenterY) + float64(i)*lineHeight
		dc.DrawStringAnchored(line, float64(DeskripsiCenterX), y, 0.5, 0.5)
	}
}

// drawRoundedRect draws a rounded rectangle path.
func drawRoundedRect(dc *gg.Context, x, y, w, h, r float64) {
	dc.NewSubPath()
	dc.MoveTo(x+r, y)
	dc.LineTo(x+w-r, y)
	dc.DrawArc(x+w-r, y+r, r, -gg.Radians(90), 0)
	dc.LineTo(x+w, y+h-r)
	dc.DrawArc(x+w-r, y+h-r, r, 0, gg.Radians(90))
	dc.LineTo(x+r, y+h)
	dc.DrawArc(x+r, y+h-r, r, gg.Radians(90), gg.Radians(180))
	dc.LineTo(x, y+r)
	dc.DrawArc(x+r, y+r, r, gg.Radians(180), gg.Radians(270))
	dc.ClosePath()
}

// loadImage loads an image from file.
func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}
