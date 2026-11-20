// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

// Package domain contains the domain types that feed cofacture's builders.
// They are plain structs with no persistence concerns (no db/gorm tags, no SQL) — the
// equivalent of a DTO.
//
// These types do not validate DIAN catalog codes (document type, tax, payment method, etc.)
// — they receive codes that have already been validated by whoever constructs them (the
// orchestrator). The core only knows which XML node a code belongs in, not whether the code
// itself is valid.
package domain

import "github.com/movaltech/decimal"

// Identification represents a third party's identification number (NIT, national ID, etc.)
// along with the attributes DIAN requires in the schemeID/schemeName schemas.
type Identification struct {
	Number           string // e.g. "900123456"
	TypeCode         string // identification_types catalog, e.g. "31" = NIT
	VerificationCode string // check digit, only applies when TypeCode == "31"
}

// Address is the physical/registered address of a Party.
type Address struct {
	Line        string
	CityCode    string
	CityName    string
	PostalZone  string
	StateCode   string
	StateName   string
	CountryCode string
	CountryName string
}

// Party is a supplier or customer (AccountingSupplierParty / AccountingCustomerParty).
type Party struct {
	// EntityTypeCode is cbc:AdditionalAccountID, DIAN table 13.2.3 (TipoOrganizacion): "1" legal
	// entity (persona jurídica), "2" natural person (persona natural). Technical Annex 1.9 rule
	// FAK19 states it the same way.
	EntityTypeCode string
	Identification Identification
	Name           string
	Address        Address // omitted from the XML if Line is empty
	TaxRegimeCode  string  // type_regimes catalog, used as the listName of TaxLevelCode
	// LiabilityCodes are the tax liability responsibilities (tax_level_codes catalog,
	// e.g. O-13, O-15, O-47). DIAN allows more than one per third party — they are
	// serialized joined with ";" in a single cbc:TaxLevelCode, not one element per code.
	LiabilityCodes []string
	// IndustryClassificationCodes are the CIIU codes (DANE catalog, not DIAN — hence there is
	// no catalog table validating them; any cardinality limit is up to whoever builds the
	// Party). Only applies to the supplier, never to the customer (the Technical Annex
	// describes it as "the issuer's economic activity code"). Serialized joined with ";" in a
	// single cbc:IndustryClassificationCode.
	IndustryClassificationCodes []string
	TaxSchemeCode               string
	TaxSchemeName               string
	Phone                       string
	Email                       string
	// MerchantRegistrationNumber is the Chamber of Commerce ("Cámara de Comercio")
	// registration number. nil if not applicable (always nil for the customer; optional for
	// the supplier — e.g. a natural person without a mercantile registration won't have one).
	// The cbc:ID of CorporateRegistrationScheme is NOT this number: it is the invoice prefix.
	MerchantRegistrationNumber *string
}

// Tax is a tax subtotal, at either line or header level.
type Tax struct {
	TaxableAmountCents int64
	TaxAmountCents     int64
	Percent            decimal.Decimal // e.g. decimal.MustParse("19"); rendered with 2 decimals
	TypeCode           string          // tax_types catalog, e.g. "01" = VAT
	TypeName           string
}

// ReferencePrice is used for free samples or no-charge items (PricingReference).
type ReferencePrice struct {
	PriceAmountCents int64
	TypeCode         string
}

// AllowanceCharge is a discount (ChargeIndicator false) or a charge (true), cac:AllowanceCharge.
// On a Line it changes the line's taxable base: the caller subtracts discounts from and adds
// charges to Line.LineExtensionCents (rule FAV06). On the Document it does not touch any tax
// base (rule FAQ01) and only moves PayableCents (rule FAU14).
type AllowanceCharge struct {
	ChargeIndicator bool
	// ReasonCode is table 13.3.8: "00"/"01" discounts, "02"/"03" charges. DIAN requires it
	// for a document-level discount (rule FAQ04).
	ReasonCode      string
	Reason          string          // free text, e.g. "Descuento por pronto pago"
	Percent         decimal.Decimal // cbc:MultiplierFactorNumeric, 0-100, rendered with 2 decimals
	AmountCents     int64
	BaseAmountCents int64
}

// PrepaidPayment is an advance payment received before the invoice (cac:PrepaidPayment, group
// FBD01). Its total goes to Totals.PrepaidCents; DIAN does not subtract it from PayableCents
// (rule FAU14).
type PrepaidPayment struct {
	ID              string
	PaidAmountCents int64
	ReceivedDate    string // YYYY-MM-DD
	PaidDate        string // YYYY-MM-DD, optional
}

// Line is a detail line (InvoiceLine / CreditNoteLine / DebitNoteLine).
type Line struct {
	Description        string
	Quantity           decimal.Decimal // rendered with 6 decimals
	UnitCode           string          // unit_codes catalog, e.g. "94"
	LineExtensionCents int64
	UnitPriceCents     int64
	FreeOfCharge       bool
	ReferencePrice     *ReferencePrice // required when FreeOfCharge == true
	ItemCode           string
	ItemTypeCode       string // product coding standard catalog (e.g. "999")
	ItemTypeName       string
	ItemTypeAgencyID   string
	// AllowanceCharges are the line's discounts and charges, already reflected in
	// LineExtensionCents. Ignored by the Support Document and its Adjustment Note.
	AllowanceCharges []AllowanceCharge
	Taxes            []Tax
}

// PaymentMean is a payment method (PaymentMeans).
type PaymentMean struct {
	Code              string // payment term: cash/credit (the orchestrator's "payment_terms" catalog)
	PaymentMethodCode string // the payment method itself: cash, transfer... (the orchestrator's "payment_methods" catalog)
	DueDate           string // only applies when Code == "2" (credit)
	PaymentReference  string
}

// Totals are the document's legal totals (LegalMonetaryTotal). The builder package renders
// these fields as given — it does not recompute or cross-check them against Line.LineExtensionCents
// or Line.Taxes. Getting LineExtensionCents/TaxExclusiveCents/TaxInclusiveCents/PayableCents to
// actually add up to the sum of the document's lines is the caller's responsibility, same as
// every other value in this package (see the no-catalog-validation note above).
type Totals struct {
	LineExtensionCents int64
	TaxExclusiveCents  int64
	TaxInclusiveCents  int64
	// AllowanceTotalCents and ChargeTotalCents are the sums of the document-level discounts
	// and charges (rules FAU08 and FAU10). Only serialized when > 0.
	AllowanceTotalCents int64
	ChargeTotalCents    int64
	PrepaidCents        int64 // only serialized when > 0 and the document is an Invoice

	// RoundingCents is cbc:PayableRoundingAmount — the only monetary field the Technical Annex
	// allows to be negative (e.g. rounding a POS total to the nearest 50/100 pesos). Only
	// serialized when non-zero.
	RoundingCents int64

	PayableCents int64
}

// NumberingRange is the DIAN-authorized numbering resolution range that covers this document
// (InvoiceControl).
type NumberingRange struct {
	AuthorizedCode string
	Prefix         string
	StartNumber    string
	EndNumber      string
	StartDate      string // YYYY-MM-DD
	EndDate        string // YYYY-MM-DD
}

// SoftwareProvider identifies the technology provider and the software authorized by DIAN.
type SoftwareProvider struct {
	ProviderIdentification Identification
	SoftwareID             string
}
