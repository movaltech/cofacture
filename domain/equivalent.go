// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package domain

// SoftwareManufacturer is the FabricanteSoftware extension that every Documento Equivalente
// Electrónico carries (Technical Annex Documento Equivalente Electrónico V1.0, section 8.1,
// DEAB37-DEAB46): who made the software that generated the document. The builder renders it
// only when set; whether a document needs it is the caller's decision.
type SoftwareManufacturer struct {
	// Name is "NombreApellido": names and surnames of the manufacturer (required by the annex).
	Name string

	// LegalName is "RazonSocial", the legal name of the manufacturer; omitted when empty. Section
	// 8.1 and the general table of 8.2 mark it 0..1, but the POS extension table of 8.2 marks it
	// 1..1 (DEAB43/DEAB44) and DIAN rejected a POS ticket without it in habilitación: set it on
	// every POS ticket.
	LegalName string

	// SoftwareName is "NombreSoftware": name of the software (required by the annex).
	SoftwareName string
}

// POSBuyerBenefits is the BeneficiosComprador extension of the POS ticket (InvoiceTypeCode
// "20"; annex section 8.2.1.1, DEPD11-DEPD20): the buyer and the loyalty points they hold.
// The annex tables spell the elements with a V ("Veneficios"); the builder uses the B of its rule
// DEPD15 and of the published examples. To be confirmed in the DIAN test set.
type POSBuyerBenefits struct {
	Code   string // "Codigo": the buyer's identification number
	Name   string // "NombresApellidos": the buyer's names and surnames
	Points string // "Puntos": loyalty points accumulated by the buyer
}

// POSCashRegister is the PuntoVenta extension of the POS ticket (annex section 8.2.1.2,
// DEPD21-DEPD36): the cash register the sale was made on.
type POSCashRegister struct {
	Plate    string // "PlacaCaja": inventory plate of the register
	Location string // "UbicaciónCaja": where the register is
	Cashier  string // "Cajero": names and surnames of the cashier or seller
	Type     string // "TipoCaja": type of register
	SaleCode string // "CódigoVenta": code of the sale

	SubtotalCents int64 // "SubTotal": subtotal of the sale
}

// AdditionalDocumentReference is a cac:AdditionalDocumentReference (Documento Equivalente annex
// V1.0, DEAI01-DEAI06): a document this one references. Only ID is required; the empty fields
// are not rendered.
type AdditionalDocumentReference struct {
	ID               string // cbc:ID: prefix and number of the referenced document
	SchemeName       string // cbc:ID/@schemeName, e.g. the hash scheme of a referenced key
	IssueDate        string // cbc:IssueDate, YYYY-MM-DD
	DocumentTypeCode string // cbc:DocumentTypeCode, the issuer's own code
}
