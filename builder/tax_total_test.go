// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"strings"
	"testing"

	"github.com/movaltech/cofacture/v2/domain"
	"github.com/movaltech/decimal"
)

// TestTaxTotalPerTaxType: an invoice with IVA and INC carries one header TaxTotal per tax type
// with the sum of its subtotals (rules FAS01b and FAS02); DIAN rejected a single TaxTotal
// holding both.
func TestTaxTotalPerTaxType(t *testing.T) {
	inv := sampleInvoice()
	inv.HeaderTaxes = []domain.Tax{
		{TaxableAmountCents: 10000, TaxAmountCents: 1900, Percent: decimal.MustParse("19"), TypeCode: "01", TypeName: "IVA"},
		{TaxableAmountCents: 20000, TaxAmountCents: 1600, Percent: decimal.MustParse("8"), TypeCode: "04", TypeName: "INC"},
		{TaxableAmountCents: 5000, TaxAmountCents: 250, Percent: decimal.MustParse("5"), TypeCode: "01", TypeName: "IVA"},
	}
	doc, err := BuildInvoice(inv)
	if err != nil {
		t.Fatal(err)
	}
	header := doc.Root().SelectElements("cac:TaxTotal")
	if len(header) != 2 {
		t.Fatalf("header TaxTotal = %d, want one per tax type (2)", len(header))
	}
	for i, want := range []struct {
		amount    string
		subtotals int
		scheme    string
	}{{"21.50", 2, "01"}, {"16.00", 1, "04"}} {
		tt := header[i]
		if got := tt.SelectElement("cbc:TaxAmount").Text(); got != want.amount {
			t.Errorf("TaxTotal %d amount = %s, want %s", i, got, want.amount)
		}
		subs := tt.SelectElements("cac:TaxSubtotal")
		if len(subs) != want.subtotals {
			t.Errorf("TaxTotal %d subtotals = %d, want %d", i, len(subs), want.subtotals)
		}
		for _, s := range subs {
			if id := s.FindElement("cac:TaxCategory/cac:TaxScheme/cbc:ID").Text(); !strings.EqualFold(id, want.scheme) {
				t.Errorf("TaxTotal %d holds tax %s, want only %s", i, id, want.scheme)
			}
		}
	}
}
