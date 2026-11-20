// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"github.com/beevik/etree"
	"github.com/movaltech/cofacture/v2/domain"
)

// appendWithholdingTaxTotal adds a cac:WithholdingTaxTotal for each withholding. Unlike
// cac:TaxTotal, where every subtype goes under the same element, each withholding generates
// its own cac:WithholdingTaxTotal with exactly one cac:TaxSubtotal. exemptionReason adds
// cbc:TaxExemptionReason "01", which the Support Document and its Adjustment Note carry; the
// Invoice and its notes do not have that element.
func appendWithholdingTaxTotal(parent *etree.Element, taxes []domain.Tax, currency string, exemptionReason bool) {
	for _, t := range taxes {
		wtt := parent.CreateElement("cac:WithholdingTaxTotal")

		amt := wtt.CreateElement("cbc:TaxAmount")
		amt.CreateAttr("currencyID", currency)
		amt.SetText(formatAmount(t.TaxAmountCents))

		sub := wtt.CreateElement("cac:TaxSubtotal")

		taxable := sub.CreateElement("cbc:TaxableAmount")
		taxable.CreateAttr("currencyID", currency)
		taxable.SetText(formatAmount(t.TaxableAmountCents))

		taxAmt := sub.CreateElement("cbc:TaxAmount")
		taxAmt.CreateAttr("currencyID", currency)
		taxAmt.SetText(formatAmount(t.TaxAmountCents))

		category := sub.CreateElement("cac:TaxCategory")
		category.CreateElement("cbc:Percent").SetText(formatPercent(t.Percent))
		if exemptionReason {
			category.CreateElement("cbc:TaxExemptionReason").SetText("01")
		}
		scheme := category.CreateElement("cac:TaxScheme")
		scheme.CreateElement("cbc:ID").SetText(t.TypeCode)
		scheme.CreateElement("cbc:Name").SetText(t.TypeName)
	}
}

// appendTaxTotal adds cac:TaxTotal (header or line level). Adds nothing if taxes is empty,
// matching the technical annex's requirement that TaxTotal is optional when no tax applies.
//
// Each tax type gets its own cac:TaxTotal, whose TaxAmount is the sum of that type's
// subtotals (rules FAS01b and FAS02): an Invoice with IVA and INC carries two TaxTotal. Types
// keep the order of their first appearance in taxes.
func appendTaxTotal(parent *etree.Element, taxes []domain.Tax, currency string) {
	var order []string
	byType := map[string][]domain.Tax{}
	for _, t := range taxes {
		if _, ok := byType[t.TypeCode]; !ok {
			order = append(order, t.TypeCode)
		}
		byType[t.TypeCode] = append(byType[t.TypeCode], t)
	}
	for _, typ := range order {
		appendTaxTotalOfType(parent, byType[typ], currency)
	}
}

// appendTaxTotalOfType adds one cac:TaxTotal with the subtotals of a single tax type.
func appendTaxTotalOfType(parent *etree.Element, taxes []domain.Tax, currency string) {
	var totalCents int64
	for _, t := range taxes {
		totalCents += t.TaxAmountCents
	}

	taxTotal := parent.CreateElement("cac:TaxTotal")
	amt := taxTotal.CreateElement("cbc:TaxAmount")
	amt.CreateAttr("currencyID", currency)
	amt.SetText(formatAmount(totalCents))

	for _, t := range taxes {
		sub := taxTotal.CreateElement("cac:TaxSubtotal")

		taxable := sub.CreateElement("cbc:TaxableAmount")
		taxable.CreateAttr("currencyID", currency)
		taxable.SetText(formatAmount(t.TaxableAmountCents))

		taxAmt := sub.CreateElement("cbc:TaxAmount")
		taxAmt.CreateAttr("currencyID", currency)
		taxAmt.SetText(formatAmount(t.TaxAmountCents))

		category := sub.CreateElement("cac:TaxCategory")
		category.CreateElement("cbc:Percent").SetText(formatPercent(t.Percent))
		scheme := category.CreateElement("cac:TaxScheme")
		scheme.CreateElement("cbc:ID").SetText(t.TypeCode)
		scheme.CreateElement("cbc:Name").SetText(t.TypeName)
	}
}
