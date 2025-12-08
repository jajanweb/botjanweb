// Package adapters contains infrastructure adapters that implement domain ports.
// These adapters follow the Hexagonal Architecture pattern (Ports & Adapters),
// allowing the domain/use-case layer to remain independent of infrastructure details.
package adapters

import (
	infraqris "github.com/exernia/botjanweb/internal/infrastructure/external/qris"
)

// QrisGeneratorAdapter adapts infraqris.Generator to usecase.QrisGeneratorPort.
// This adapter implements the Hexagonal Architecture pattern, allowing the usecase
// layer to depend on an interface rather than a concrete infrastructure implementation.
type QrisGeneratorAdapter struct {
	gen *infraqris.Generator
}

// NewQrisGeneratorAdapter creates a new adapter instance.
func NewQrisGeneratorAdapter(gen *infraqris.Generator) *QrisGeneratorAdapter {
	return &QrisGeneratorAdapter{gen: gen}
}

// GenerateDynamicQRIS implements usecase.QrisGeneratorPort interface.
func (a *QrisGeneratorAdapter) GenerateDynamicQRIS(baseQR string, amount int, description string) (string, []byte, error) {
	result, err := a.gen.GenerateQris(baseQR, amount, description)
	if err != nil {
		return "", nil, err
	}
	return result.QrisString, result.ImageData, nil
}
