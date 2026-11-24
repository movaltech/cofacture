// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"strings"
	"testing"

	"github.com/movaltech/cofacture/v2/domain"
)

func buildString(t *testing.T, inv domain.Document) string {
	t.Helper()
	doc, err := BuildInvoice(inv)
	if err != nil {
		t.Fatalf("BuildInvoice: %v", err)
	}
	s, err := doc.WriteToString()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// A POS ticket in contingency keeps its prefix and number (annex section 15), so it keeps the
// authorized numbering when its range is given; without a range there is none to render.
func TestInvoiceControl_Contingency(t *testing.T) {
	inv := samplePOSInvoice()
	inv.DocumentTypeCode = "08"
	if got := buildString(t, inv); !strings.Contains(got, "<sts:InvoiceControl>") {
		t.Error("a type 08 document with its numbering range lost sts:InvoiceControl")
	}
	inv.NumberingRange = domain.NumberingRange{}
	if got := buildString(t, inv); strings.Contains(got, "<sts:InvoiceControl>") {
		t.Error("a type 08 document without a numbering range rendered sts:InvoiceControl")
	}
}

func TestEquivalentExtensions(t *testing.T) {
	// Only the extensions that are set are rendered, before the signature placeholder.
	inv := samplePOSInvoice()
	inv.SoftwareManufacturer.LegalName = ""
	inv.POSBuyerBenefits = nil
	got := buildString(t, inv)
	if strings.Contains(got, "RazonSocial") {
		t.Error("an empty legal name was rendered")
	}
	if strings.Contains(got, "BeneficiosComprador") {
		t.Error("buyer benefits rendered without being set")
	}
	if strings.Index(got, "<PuntoVenta>") > strings.LastIndex(got, "<ext:UBLExtension>") {
		t.Error("an extension went after the signature placeholder")
	}

	// A regular invoice carries none of them.
	if got := buildString(t, sampleInvoice()); strings.Contains(got, "FabricanteSoftware") || strings.Contains(got, "PuntoVenta") {
		t.Error("a regular invoice rendered Documento Equivalente extensions")
	}
}

func TestAdditionalDocumentReference(t *testing.T) {
	inv := samplePOSInvoice()
	if got := buildString(t, inv); strings.Contains(got, "AdditionalDocumentReference") {
		t.Error("a document without references rendered one")
	}
	inv.DocumentTypeCode = "08"
	inv.AdditionalDocumentReferences = []domain.AdditionalDocumentReference{{ID: "POS1", IssueDate: "2026-11-24"}}
	got := buildString(t, inv)
	want := "<cac:AdditionalDocumentReference><cbc:ID>POS1</cbc:ID><cbc:IssueDate>2026-11-24</cbc:IssueDate></cac:AdditionalDocumentReference>"
	if !strings.Contains(strings.Join(strings.Fields(got), ""), want) {
		t.Errorf("the reference is not rendered as %s", want)
	}
	// UBL 2.1 order: after the document's own fields, before the supplier.
	if strings.Index(got, "<cac:AdditionalDocumentReference>") > strings.Index(got, "<cac:AccountingSupplierParty>") {
		t.Error("the reference went after the supplier party")
	}
}
