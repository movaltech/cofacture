// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/movaltech/cofacture/v2/domain"
	"github.com/movaltech/decimal"
)

// The nominal subtotals below are those of DIAN's own examples in its toolbox (Caja de
// herramientas FE V1.9, v2026-2): "Ejemplicacion Impuestos Saludables.xml" for IBUA and
// "Consumidor Final.xml" for plastic bags, except that BaseUnitMeasure always has 2 decimals
// (rule FAS09).
var (
	ibua = domain.Tax{
		TaxAmountCents: 450000, TypeCode: "34", TypeName: "IBUA",
		BaseUnitCode: "MLT", BaseUnitMeasure: decimal.MustParse("25000"), PerUnitAmountCents: 1800,
	}
	bags = domain.Tax{
		TaxAmountCents: 6000, TypeCode: "22", TypeName: "Bolsas",
		BaseUnitCode: "NIU", BaseUnitMeasure: decimal.MustParse("1"), PerUnitAmountCents: 3000,
	}
)

const (
	ibuaSubtotal = `<cac:TaxSubtotal><cbc:TaxAmount currencyID="COP">4500.00</cbc:TaxAmount>` +
		`<cbc:BaseUnitMeasure unitCode="MLT">25000.00</cbc:BaseUnitMeasure>` +
		`<cbc:PerUnitAmount currencyID="COP">18.00</cbc:PerUnitAmount>` +
		`<cac:TaxCategory><cac:TaxScheme><cbc:ID>34</cbc:ID><cbc:Name>IBUA</cbc:Name></cac:TaxScheme></cac:TaxCategory></cac:TaxSubtotal>`
	bagsSubtotal = `<cac:TaxSubtotal><cbc:TaxAmount currencyID="COP">60.00</cbc:TaxAmount>` +
		`<cbc:BaseUnitMeasure unitCode="NIU">1.00</cbc:BaseUnitMeasure>` +
		`<cbc:PerUnitAmount currencyID="COP">30.00</cbc:PerUnitAmount>` +
		`<cac:TaxCategory><cac:TaxScheme><cbc:ID>22</cbc:ID><cbc:Name>Bolsas</cbc:Name></cac:TaxScheme></cac:TaxCategory></cac:TaxSubtotal>`
)

// TestBuild_NominalTaxes checks that IBUA and the plastic bag tax render as nominal subtotals,
// at header and line level, next to a percentage tax that renders as before, on an invoice, a
// POS ticket and a POS credit note.
func TestBuild_NominalTaxes(t *testing.T) {
	withNominal := func(doc domain.Document) domain.Document {
		doc.HeaderTaxes = append(append([]domain.Tax{}, doc.HeaderTaxes...), ibua, bags)
		lines := append([]domain.Line{}, doc.Lines...)
		lines[0].Taxes = append(append([]domain.Tax{}, lines[0].Taxes...), ibua)
		doc.Lines = lines
		return doc
	}
	build := map[string]func() (*etree.Document, error){
		"invoice":    func() (*etree.Document, error) { return BuildInvoice(withNominal(sampleInvoice())) },
		"POS ticket": func() (*etree.Document, error) { return BuildInvoice(withNominal(samplePOSInvoice())) },
		"POS credit note": func() (*etree.Document, error) {
			cn := samplePOSCreditAdjustment()
			cn.Document = withNominal(cn.Document)
			return BuildCreditNote(cn)
		},
	}

	for name, b := range build {
		tree, err := b()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		xml, err := tree.WriteToString()
		if err != nil {
			t.Fatal(err)
		}
		flat := compact(xml)

		// Header: IBUA and bags, each in its own TaxTotal (one per tax type).
		for _, want := range []string{
			`<cac:TaxTotal><cbc:TaxAmount currencyID="COP">4500.00</cbc:TaxAmount>` + ibuaSubtotal + `</cac:TaxTotal>`,
			`<cac:TaxTotal><cbc:TaxAmount currencyID="COP">60.00</cbc:TaxAmount>` + bagsSubtotal + `</cac:TaxTotal>`,
		} {
			if !strings.Contains(flat, want) {
				t.Errorf("%s: header does not contain %s", name, want)
			}
		}
		// Line: IBUA next to the line's percentage tax.
		line := flat[strings.Index(flat, "Line><cbc:ID>1</cbc:ID>"):]
		if !strings.Contains(line, ibuaSubtotal) {
			t.Errorf("%s: line 1 does not contain the IBUA subtotal", name)
		}
		// The percentage taxes keep TaxableAmount and Percent.
		if !strings.Contains(flat, `<cac:TaxSubtotal><cbc:TaxableAmount`) || !strings.Contains(flat, `<cbc:Percent>`) {
			t.Errorf("%s: the percentage taxes lost TaxableAmount or Percent", name)
		}
	}
}

// TestBuild_NominalTaxWithTaxableAmount checks that a nominal tax carries cbc:TaxableAmount
// when the caller sets it, before cbc:TaxAmount as UBL 2.1 orders them, and still no Percent.
func TestBuild_NominalTaxWithTaxableAmount(t *testing.T) {
	doc := sampleInvoice()
	tax := bags
	tax.TaxableAmountCents = 200
	doc.HeaderTaxes = []domain.Tax{tax}

	tree, err := BuildInvoice(doc)
	if err != nil {
		t.Fatal(err)
	}
	xml, err := tree.WriteToString()
	if err != nil {
		t.Fatal(err)
	}
	want := `<cac:TaxSubtotal><cbc:TaxableAmount currencyID="COP">2.00</cbc:TaxableAmount><cbc:TaxAmount currencyID="COP">60.00</cbc:TaxAmount><cbc:BaseUnitMeasure unitCode="NIU">1.00</cbc:BaseUnitMeasure><cbc:PerUnitAmount currencyID="COP">30.00</cbc:PerUnitAmount><cac:TaxCategory><cac:TaxScheme>`
	if !strings.Contains(compact(xml), want) {
		t.Errorf("XML does not contain %s", want)
	}
}
