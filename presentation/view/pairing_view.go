// Package view provides presentation layer rendering using modern builder pattern.
// This package follows Clean Architecture by separating view logic from HTTP handlers.
package view

import (
	"html/template"
	"net/http"

	"github.com/exernia/botjanweb/presentation/view/html"
)

// PairingView handles rendering of WhatsApp pairing pages.
// Uses builder pattern for flexible, testable, and scalable view rendering.
type PairingView struct {
	successTemplate *template.Template
	pageTemplate    *template.Template
}

// NewPairingView creates a new pairing view renderer.
// Templates are parsed once at initialization for performance.
func NewPairingView() (*PairingView, error) {
	successTmpl, err := template.New("success").Parse(html.PairingSuccessTemplate)
	if err != nil {
		return nil, err
	}

	pageTmpl, err := template.New("page").Parse(html.PairingPageTemplate)
	if err != nil {
		return nil, err
	}

	return &PairingView{
		successTemplate: successTmpl,
		pageTemplate:    pageTmpl,
	}, nil
}

// PairingSuccessData contains data for the success page view.
type PairingSuccessData struct {
	DeviceID string
}

// PairingPageData contains data for the pairing page view.
type PairingPageData struct {
	Token string
}

// RenderSuccess renders the "Already Connected" success page.
// This page is shown when WhatsApp is already paired and logged in.
func (v *PairingView) RenderSuccess(w http.ResponseWriter, data PairingSuccessData) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return v.successTemplate.Execute(w, data)
}

// RenderPage renders the interactive QR code pairing page.
// This page includes QR polling, instructions, and status indicators.
func (v *PairingView) RenderPage(w http.ResponseWriter, data PairingPageData) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return v.pageTemplate.Execute(w, data)
}
