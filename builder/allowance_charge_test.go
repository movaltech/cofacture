// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"strings"
	"testing"

	"github.com/movaltech/cofacture/v2/domain"
	"github.com/movaltech/decimal"
)

// TestBuildInvoice_AllowancesPrepaidWithholdings checks the elements of discounts, charges,
// advance payments and withholdings, and that they appear in the order UBL 2.1 gives them.
func TestBuildInvoice_AllowancesPrepaidWithholdings(t *testing.T) {
	inv := sampleInvoice()
	inv.Lines[0].AllowanceCharges = []domain.AllowanceCharge{{
		Reason: "Descuento por volumen", Percent: decimal.MustParse("10"), AmountCents: 1000, BaseAmountCents: 10000,
	}}
	inv.PrepaidPayments = []domain.PrepaidPayment{{ID: "ANT-1", PaidAmountCents: 5000, ReceivedDate: "2026-11-01"}}
	inv.AllowanceCharges = []domain.AllowanceCharge{
		{ReasonCode: "00", Reason: "Descuento comercial", Percent: decimal.MustParse("5"), AmountCents: 500, BaseAmountCents: 10000},
		{ChargeIndicator: true, ReasonCode: "03", Reason: "Propina", Percent: decimal.MustParse("10"), AmountCents: 1000, BaseAmountCents: 10000},
	}
	inv.WithholdingTaxes = []domain.Tax{{TaxableAmountCents: 10000, TaxAmountCents: 250, Percent: decimal.MustParse("2.5"), TypeCode: "06", TypeName: "ReteFuente"}}
	inv.Totals.AllowanceTotalCents, inv.Totals.ChargeTotalCents, inv.Totals.PrepaidCents = 500, 1000, 5000

	doc, err := BuildInvoice(inv)
	if err != nil {
		t.Fatal(err)
	}
	xml, err := doc.WriteToString()
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"<cbc:ID>ANT-1</cbc:ID><cbc:PaidAmount currencyID=\"COP\">50.00</cbc:PaidAmount><cbc:ReceivedDate>2026-11-01</cbc:ReceivedDate>",
		"<cbc:ID>1</cbc:ID><cbc:ChargeIndicator>false</cbc:ChargeIndicator><cbc:AllowanceChargeReasonCode>00</cbc:AllowanceChargeReasonCode>",
		"<cbc:ID>2</cbc:ID><cbc:ChargeIndicator>true</cbc:ChargeIndicator><cbc:AllowanceChargeReasonCode>03</cbc:AllowanceChargeReasonCode>",
		"<cbc:MultiplierFactorNumeric>10.00</cbc:MultiplierFactorNumeric>",
		"<cbc:AllowanceTotalAmount currencyID=\"COP\">5.00</cbc:AllowanceTotalAmount><cbc:ChargeTotalAmount currencyID=\"COP\">10.00</cbc:ChargeTotalAmount><cbc:PrepaidAmount currencyID=\"COP\">50.00</cbc:PrepaidAmount>",
		"<cbc:Name>ReteFuente</cbc:Name>",
		"<cbc:Percent>2.50</cbc:Percent>",
	} {
		if !strings.Contains(compact(xml), want) {
			t.Errorf("XML does not contain %s", want)
		}
	}
	if strings.Contains(xml, "TaxExemptionReason") {
		t.Error("an invoice withholding must not carry TaxExemptionReason")
	}

	// UBL 2.1 order at document level, and inside the line.
	order := []string{"<cac:PaymentMeans>", "<cac:PrepaidPayment>", "<cac:AllowanceCharge>", "<cac:TaxTotal>", "<cac:WithholdingTaxTotal>", "<cac:LegalMonetaryTotal>", "<cac:InvoiceLine>"}
	assertOrder(t, xml, order)
	line := xml[strings.Index(xml, "<cac:InvoiceLine>"):]
	assertOrder(t, line, []string{"<cbc:LineExtensionAmount", "<cac:AllowanceCharge>", "<cac:TaxTotal>", "<cac:Item>"})
}

// TestBuild_PrepaidAmountByBuilder checks that cbc:PrepaidAmount follows the builder, not the
// type code: BuildInvoice renders it for an invoice and a POS ticket (Documento Equivalente
// annex DEBD01, DEAU12), and the notes still leave it out.
func TestBuild_PrepaidAmountByBuilder(t *testing.T) {
	const prepaidAmount = `<cbc:PrepaidAmount currencyID="COP">50.00</cbc:PrepaidAmount>`
	withPrepaid := func(doc domain.Document) domain.Document {
		doc.PrepaidPayments = []domain.PrepaidPayment{{ID: "ANT-1", PaidAmountCents: 5000, ReceivedDate: "2026-11-01"}}
		doc.Totals.PrepaidCents = 5000
		return doc
	}

	for name, doc := range map[string]domain.Document{"invoice": sampleInvoice(), "POS ticket": samplePOSInvoice()} {
		tree, err := BuildInvoice(withPrepaid(doc))
		if err != nil {
			t.Fatal(err)
		}
		xml, err := tree.WriteToString()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(compact(xml), prepaidAmount) {
			t.Errorf("%s: missing %s", name, prepaidAmount)
		}
		assertOrder(t, xml, []string{"<cac:PrepaidPayment>", "<cac:LegalMonetaryTotal>", "<cbc:PrepaidAmount", "<cbc:PayableAmount"})
	}

	cn := samplePOSCreditAdjustment()
	cn.Document = withPrepaid(cn.Document)
	tree, err := BuildCreditNote(cn)
	if err != nil {
		t.Fatal(err)
	}
	xml, err := tree.WriteToString()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(xml, "Prepaid") {
		t.Error("a credit note renders no advance payments")
	}
}

func compact(xml string) string {
	var b strings.Builder
	for _, l := range strings.Split(xml, "\n") {
		b.WriteString(strings.TrimSpace(l))
	}
	return b.String()
}

func assertOrder(t *testing.T, xml string, elements []string) {
	t.Helper()
	last := -1
	for _, e := range elements {
		i := strings.Index(xml, e)
		if i < 0 || i < last {
			t.Errorf("%s is missing or out of order", e)
		}
		last = i
	}
}

// TestBuildNotes_AllowanceChargeOrder checks the positions the Credit and Debit Note schemas
// give AllowanceCharge, which differ from the Invoice's (DIAN rejects other orders with ZB01).
// The adjustment notes to an equivalent document ("94" credit, "93" debit) use the same schemas
// and must follow the same order as the invoice notes ("91", "92").
func TestBuildNotes_AllowanceChargeOrder(t *testing.T) {
	base := sampleInvoice()
	base.Lines[0].AllowanceCharges = []domain.AllowanceCharge{{Reason: "Descuento", Percent: decimal.MustParse("5"), AmountCents: 500, BaseAmountCents: 10000}}
	base.AllowanceCharges = []domain.AllowanceCharge{{ChargeIndicator: true, ReasonCode: "02", Reason: "Recargo", Percent: decimal.MustParse("1"), AmountCents: 100, BaseAmountCents: 10000}}
	base.WithholdingTaxes = []domain.Tax{{TaxableAmountCents: 10000, TaxAmountCents: 250, Percent: decimal.MustParse("2.5"), TypeCode: "06", TypeName: "ReteFuente"}}
	ref := domain.BillingReference{Prefix: "SETP", Number: "1", CUFE: "abc", IssueDate: "2026-11-19", HashType: "CUFE-SHA384"}

	for _, codes := range []struct{ credit, debit string }{{"91", "92"}, {"94", "93"}} {
		cnDoc := base
		cnDoc.DocumentTypeCode = codes.credit
		cn, err := BuildCreditNote(domain.CreditNote{Document: cnDoc, CreditNoteTypeCode: codes.credit, BillingReference: ref})
		if err != nil {
			t.Fatal(err)
		}
		dnDoc := base
		dnDoc.DocumentTypeCode = codes.debit
		dn, err := BuildDebitNote(domain.DebitNote{Document: dnDoc, BillingReference: ref})
		if err != nil {
			t.Fatal(err)
		}
		for name, tree := range map[string]interface{ WriteToString() (string, error) }{"credit note " + codes.credit: cn, "debit note " + codes.debit: dn} {
			xml, err := tree.WriteToString()
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(xml, "WithholdingTaxTotal") {
				t.Errorf("%s: its schema has no WithholdingTaxTotal", name)
			}
			isDebit := strings.HasPrefix(name, "debit")
			lineNode := "<cac:CreditNoteLine>"
			if isDebit {
				lineNode = "<cac:DebitNoteLine>"
			}
			line := xml[strings.Index(xml, lineNode):]
			assertOrder(t, line, []string{"<cac:TaxTotal>", "<cac:AllowanceCharge>", "<cac:Item>"})
			header := xml[:strings.Index(xml, lineNode)]
			if isDebit {
				assertOrder(t, header, []string{"<cac:AccountingCustomerParty>", "<cac:AllowanceCharge>", "<cac:PaymentMeans>", "<cac:TaxTotal>"})
			} else {
				assertOrder(t, header, []string{"<cac:PaymentMeans>", "<cac:AllowanceCharge>", "<cac:TaxTotal>"})
			}
		}
	}
}
