// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"github.com/beevik/etree"
	"github.com/movaltech/cofacture/v2/domain"
)

// appendMonetaryTotal adds the totals block — shared by Invoice, CreditNote and DebitNote.
// node is "LegalMonetaryTotal" for Invoice/CreditNote, "RequestedMonetaryTotal" for DebitNote
// (same content, different element name — verified in section 8.4 of the technical annex).
//
// prepaid adds cbc:PrepaidAmount when t.PrepaidCents > 0. BuildInvoice passes true, whatever
// the document type: the invoice annex (FBD01) and the Documento Equivalente annex (DEBD01,
// DEAU12) give the Invoice its advance payments. The other builders pass false, as they render
// no cac:PrepaidPayment.
func appendMonetaryTotal(parent *etree.Element, node string, t domain.Totals, currency string, prepaid bool) {
	el := parent.CreateElement("cac:" + node)

	line := el.CreateElement("cbc:LineExtensionAmount")
	line.CreateAttr("currencyID", currency)
	line.SetText(formatAmount(t.LineExtensionCents))

	taxExcl := el.CreateElement("cbc:TaxExclusiveAmount")
	taxExcl.CreateAttr("currencyID", currency)
	taxExcl.SetText(formatAmount(t.TaxExclusiveCents))

	taxIncl := el.CreateElement("cbc:TaxInclusiveAmount")
	taxIncl.CreateAttr("currencyID", currency)
	taxIncl.SetText(formatAmount(t.TaxInclusiveCents))

	if t.AllowanceTotalCents > 0 {
		allowance := el.CreateElement("cbc:AllowanceTotalAmount")
		allowance.CreateAttr("currencyID", currency)
		allowance.SetText(formatAmount(t.AllowanceTotalCents))
	}
	if t.ChargeTotalCents > 0 {
		charge := el.CreateElement("cbc:ChargeTotalAmount")
		charge.CreateAttr("currencyID", currency)
		charge.SetText(formatAmount(t.ChargeTotalCents))
	}

	if prepaid && t.PrepaidCents > 0 {
		prepaid := el.CreateElement("cbc:PrepaidAmount")
		prepaid.CreateAttr("currencyID", currency)
		prepaid.SetText(formatAmount(t.PrepaidCents))
	}

	// PayableRoundingAmount is the only monetary field the annex allows negative (e.g. a POS
	// total rounded to the nearest 50/100 pesos) — omitted entirely when zero, same as Prepaid.
	if t.RoundingCents != 0 {
		rounding := el.CreateElement("cbc:PayableRoundingAmount")
		rounding.CreateAttr("currencyID", currency)
		rounding.SetText(formatAmount(t.RoundingCents))
	}

	payable := el.CreateElement("cbc:PayableAmount")
	payable.CreateAttr("currencyID", currency)
	payable.SetText(formatAmount(t.PayableCents))
}
