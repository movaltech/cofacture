// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

// Package qr builds what DIAN asks for in a document's QR code: the plain search URL that goes
// in sts:QRCode and, for the documents whose annex defines one, the text block of the QR printed
// on the graphic representation.
package qr

import (
	"strings"

	"github.com/movaltech/cofacture/v2/domain"
)

const (
	habilitacionBaseURL = "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr"
	produccionBaseURL   = "https://catalogo-vpfe.dian.gov.co/document/searchqr"
)

// URL builds the QR URL from the CUFE (or CUDE) and the environment code ("1" production, "2"
// certification/testing — the same value as cbc:ProfileExecutionID). DIAN uses different
// domains per environment; confirmed against two real production invoices and the technical
// annex text.
func URL(environmentCode, documentKey string) string {
	base := produccionBaseURL
	if environmentCode == "2" {
		base = habilitacionBaseURL
	}
	return base + "?documentkey=" + documentKey
}

// SupportDocumentURL builds the Support Document's QR URL.
// Uses the same searchqr endpoint as Invoice/CreditNote/DebitNote (verified: FindDocument does
// not redirect).
func SupportDocumentURL(environmentCode, cuds string) string {
	return URL(environmentCode, cuds)
}

// SupportDocumentText builds the text of the QR printed on a support document's graphic
// representation (InvoiceTypeCode "05"), as defined in section 12.2.1 of the support document
// annex (Resolution 000167 of 2021):
//
//	NumDS: <cbc:ID>
//	FecDS: <cbc:IssueDate>
//	HorDS: <cbc:IssueTime>
//	NumSNO: <seller's identification, without check digit>
//	DocAdq: <acquirer's identification, without check digit>
//	ValDS: <LineExtensionAmount>
//	ValIva: <VAT amount (tax 01)>
//	ValTolDS: <PayableAmount>
//	CUDS: <cuds>
//	QRCode: <search URL>
//
// Roles are those of the support document: Supplier is the seller not obliged to invoice
// (SNO), Customer is the acquirer issuing the document. The annex's own example names the
// acquirer NITABS; this follows the field definition, DocAdq. The text never carries the
// software PIN, and it is not what goes in sts:QRCode: that is SupportDocumentURL.
func SupportDocumentText(doc domain.Document, cuds string) string {
	return text(doc, cuds, "DS")
}

// AdjustmentNoteText builds the text of the QR printed on the graphic representation of an
// adjustment note to a support document (CreditNote, type "95"), as defined in section 12.2.2
// of the support document annex: the same fields as SupportDocumentText, named NumNAS, FecNAS,
// HorNAS, ValNAS and ValTolNAS. As there, it never carries the software PIN, and sts:QRCode
// takes the plain URL.
func AdjustmentNoteText(doc domain.Document, cuds string) string {
	return text(doc, cuds, "NAS")
}

// text writes the printed QR block of annex 12.2; suffix names the document (DS or NAS).
func text(doc domain.Document, cuds, suffix string) string {
	var vat int64
	for _, t := range doc.HeaderTaxes {
		if t.TypeCode == "01" {
			vat += t.TaxAmountCents
		}
	}
	var b strings.Builder
	line := func(key, value string) {
		b.WriteString(key)
		b.WriteString(": ")
		b.WriteString(value)
		b.WriteByte('\n')
	}
	line("Num"+suffix, doc.Prefix+doc.Number)
	line("Fec"+suffix, doc.IssueDate)
	line("Hor"+suffix, doc.IssueTime)
	line("NumSNO", doc.Supplier.Identification.Number)
	line("DocAdq", doc.Customer.Identification.Number)
	line("Val"+suffix, domain.FormatCents(doc.Totals.LineExtensionCents))
	line("ValIva", domain.FormatCents(vat))
	line("ValTol"+suffix, domain.FormatCents(doc.Totals.PayableCents))
	line("CUDS", cuds)
	b.WriteString("QRCode: ")
	b.WriteString(URL(doc.EnvironmentCode, cuds))
	return b.String()
}

// SupportDocumentContent returns SupportDocumentText; softwarePIN is ignored.
//
// Deprecated: it used to put the software PIN, which is secret, in the QR text, and its field
// names did not follow the annex. Use SupportDocumentText for the printed QR and
// SupportDocumentURL for sts:QRCode.
func SupportDocumentContent(inv domain.Document, cuds, softwarePIN string) string {
	return SupportDocumentText(inv, cuds)
}

// AdjustmentNoteContent returns AdjustmentNoteText; softwarePIN is ignored.
//
// Deprecated: it used to put the software PIN, which is secret, in the QR text, and its field
// names did not follow the annex. Use AdjustmentNoteText for the printed QR and URL for
// sts:QRCode.
func AdjustmentNoteContent(inv domain.Document, cuds, softwarePIN string) string {
	return AdjustmentNoteText(inv, cuds)
}
