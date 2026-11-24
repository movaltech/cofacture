// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package domain

// Document is the shared model for every UBL 2.1 commercial document this library builds:
// Electronic Sales Invoice (InvoiceTypeCode "01"), Support Document ("05"), Documento
// Equivalente Electrónico ("20" and its other sub-types), and — embedded — the base of
// CreditNote, DebitNote and AdjustmentNote (see notes.go). DIAN's technical annexes define
// these as ~90% the same fields (parties, lines, totals, taxes); which document a given value
// actually becomes depends entirely on which builder.Build* function you pass it to and which
// fields you set (DocumentTypeCode, HashType, inverted Supplier/Customer roles for Support
// Document, etc.) — this type does not encode or enforce that by itself.
//
// CUFE, SoftwareSecurityCode and QRURL are left empty when building the document: they are
// computed in later pipeline steps (cufe/cude/cuds, qr) from these same fields and injected
// into the already-built XML before signing.
type Document struct {
	ProfileID         string // "DIAN 2.1: Factura Electrónica de Venta"
	EnvironmentCode   string // "1" production, "2" test/certification (also used as the UUID's schemeID)
	OperationTypeCode string // operation type catalog, e.g. "10" = Standard
	DocumentTypeCode  string // "01" national sales invoice
	HashType          string // "CUFE-SHA384", used as the UUID's schemeName

	Prefix string
	Number string

	IssueDate string // YYYY-MM-DD
	IssueTime string // HH:MM:SS-05:00
	DueDate   string // optional

	// PeriodStartDate/PeriodEndDate are the acquisition period of the Support Document
	// (cac:InvoicePeriod). Empty → the builder falls back to IssueDate for both.
	PeriodStartDate string // YYYY-MM-DD
	PeriodEndDate   string // YYYY-MM-DD

	Note string // optional

	CurrencyCode string // currency_codes catalog, ISO 4217, e.g. "COP"

	OrderReferenceNumber string // optional

	// AdditionalDocumentReferences are other documents this one references
	// (cac:AdditionalDocumentReference), rendered in order. The Documento Equivalente annex
	// V1.0 requires one on a contingency document of type 08 (rule DEAI01); what it references
	// is the caller's decision.
	AdditionalDocumentReferences []AdditionalDocumentReference

	Supplier Party
	Customer Party

	PaymentMeans []PaymentMean

	// HeaderTaxes are the header-level TaxTotal entries (one per tax type, aggregating all
	// lines). Computing these totals is the responsibility of whoever builds the model, not
	// the builder package.
	HeaderTaxes []Tax

	// WithholdingTaxes are withholdings (WithholdingTaxTotal), one per type (ReteIVA "05",
	// ReteFuente "06", ReteICA "07"). Each element generates an independent
	// cac:WithholdingTaxTotal. On the Support Document they are part of the document; on an
	// Invoice they are the issuer's self-withholdings (autorretenciones, rule FAT01),
	// informative, and change no total. CreditNote and DebitNote ignore them: their UBL 2.1
	// schemas have no WithholdingTaxTotal. DIAN checks TaxAmount against TaxableAmount
	// x Percent with the rate at 2 decimals (notice FAT07 otherwise).
	WithholdingTaxes []Tax

	// AllowanceCharges are document-level discounts and charges, which do not change any tax
	// base; their sums go to Totals.AllowanceTotalCents and Totals.ChargeTotalCents.
	AllowanceCharges []AllowanceCharge

	// PrepaidPayments are advance payments (Invoice only); their sum goes to
	// Totals.PrepaidCents.
	PrepaidPayments []PrepaidPayment

	Totals Totals
	Lines  []Line

	NumberingRange   NumberingRange
	SoftwareProvider SoftwareProvider

	// SoftwareManufacturer, POSBuyerBenefits and POSCashRegister are the extensions of the
	// Documento Equivalente Electrónico (see equivalent.go). Each is rendered only when set.
	SoftwareManufacturer *SoftwareManufacturer
	POSBuyerBenefits     *POSBuyerBenefits
	POSCashRegister      *POSCashRegister

	CUFE                 string
	SoftwareSecurityCode string
	QRURL                string
}
