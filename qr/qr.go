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

// InvoiceText builds the text of the QR printed on an invoice's graphic representation, as
// defined in section 11.7 of the invoicing annex (version 1.9):
//
//	NumFac: <cbc:ID>
//	FecFac: <cbc:IssueDate>
//	HorFac: <cbc:IssueTime>
//	NitFac: <issuer's identification, without check digit>
//	DocAdq: <acquirer's identification, without check digit>
//	ValFac: <LineExtensionAmount>
//	ValIva: <VAT amount (tax 01)>
//	ValOtroIm: <sum of the other taxes>
//	ValTolFac: <PayableAmount>
//	CUFE: <cufe>
//	QRCode: <search URL>
//
// Withholdings are not taxes of the document and stay out. As for every document, sts:QRCode
// takes the plain URL, not this text.
func InvoiceText(doc domain.Document, cufe string) string {
	return invoiceText(doc, "CUFE", cufe)
}

// NoteText builds the printed QR text of a credit or debit note: the fields of InvoiceText
// (the annex uses the same names for every document of section 11.7), with the note's CUDE.
func NoteText(doc domain.Document, cude string) string {
	return invoiceText(doc, "CUDE", cude)
}

func invoiceText(doc domain.Document, keyName, key string) string {
	vat, others := taxes(doc)
	return block(doc.EnvironmentCode, key, [][2]string{
		{"NumFac", doc.Prefix + doc.Number},
		{"FecFac", doc.IssueDate},
		{"HorFac", doc.IssueTime},
		{"NitFac", doc.Supplier.Identification.Number},
		{"DocAdq", doc.Customer.Identification.Number},
		{"ValFac", domain.FormatCents(doc.Totals.LineExtensionCents)},
		{"ValIva", domain.FormatCents(vat)},
		{"ValOtroIm", domain.FormatCents(others)},
		{"ValTolFac", domain.FormatCents(doc.Totals.PayableCents)},
		{keyName, key},
	})
}

// text writes the printed QR block of annex 12.2; suffix names the document (DS or NAS).
func text(doc domain.Document, cuds, suffix string) string {
	vat, _ := taxes(doc)
	return block(doc.EnvironmentCode, cuds, [][2]string{
		{"Num" + suffix, doc.Prefix + doc.Number},
		{"Fec" + suffix, doc.IssueDate},
		{"Hor" + suffix, doc.IssueTime},
		{"NumSNO", doc.Supplier.Identification.Number},
		{"DocAdq", doc.Customer.Identification.Number},
		{"Val" + suffix, domain.FormatCents(doc.Totals.LineExtensionCents)},
		{"ValIva", domain.FormatCents(vat)},
		{"ValTol" + suffix, domain.FormatCents(doc.Totals.PayableCents)},
		{"CUDS", cuds},
	})
}

// taxes splits the header taxes into IVA (tax 01) and every other tax.
func taxes(doc domain.Document) (vat, others int64) {
	for _, t := range doc.HeaderTaxes {
		if t.TypeCode == "01" {
			vat += t.TaxAmountCents
		} else {
			others += t.TaxAmountCents
		}
	}
	return vat, others
}

// block writes "Key: value" lines and ends with the QRCode line and the search URL.
func block(environmentCode, key string, fields [][2]string) string {
	var b strings.Builder
	for _, f := range fields {
		b.WriteString(f[0])
		b.WriteString(": ")
		b.WriteString(f[1])
		b.WriteByte('\n')
	}
	b.WriteString("QRCode: ")
	b.WriteString(URL(environmentCode, key))
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
