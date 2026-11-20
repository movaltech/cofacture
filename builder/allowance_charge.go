// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"strconv"

	"github.com/beevik/etree"
	"github.com/movaltech/cofacture/v2/domain"
)

// appendAllowanceCharges adds one cac:AllowanceCharge per element, numbered from 1 (rule
// FAQ02: consecutive IDs). Used at document level (after PrepaidPayment, before TaxTotal) and
// inside a line (before its TaxTotal), the positions UBL 2.1 gives the element.
func appendAllowanceCharges(parent *etree.Element, acs []domain.AllowanceCharge, currency string) {
	for i, ac := range acs {
		el := parent.CreateElement("cac:AllowanceCharge")
		el.CreateElement("cbc:ID").SetText(strconv.Itoa(i + 1))
		el.CreateElement("cbc:ChargeIndicator").SetText(formatBool(ac.ChargeIndicator))
		if ac.ReasonCode != "" {
			el.CreateElement("cbc:AllowanceChargeReasonCode").SetText(ac.ReasonCode)
		}
		if ac.Reason != "" {
			el.CreateElement("cbc:AllowanceChargeReason").SetText(ac.Reason)
		}
		el.CreateElement("cbc:MultiplierFactorNumeric").SetText(formatPercent(ac.Percent))

		amount := el.CreateElement("cbc:Amount")
		amount.CreateAttr("currencyID", currency)
		amount.SetText(formatAmount(ac.AmountCents))

		base := el.CreateElement("cbc:BaseAmount")
		base.CreateAttr("currencyID", currency)
		base.SetText(formatAmount(ac.BaseAmountCents))
	}
}

// appendPrepaidPayments adds one cac:PrepaidPayment per advance payment (group FBD01). Only an
// Invoice carries them.
func appendPrepaidPayments(parent *etree.Element, payments []domain.PrepaidPayment, currency string) {
	for _, p := range payments {
		el := parent.CreateElement("cac:PrepaidPayment")
		el.CreateElement("cbc:ID").SetText(p.ID)

		paid := el.CreateElement("cbc:PaidAmount")
		paid.CreateAttr("currencyID", currency)
		paid.SetText(formatAmount(p.PaidAmountCents))

		el.CreateElement("cbc:ReceivedDate").SetText(p.ReceivedDate)
		if p.PaidDate != "" {
			el.CreateElement("cbc:PaidDate").SetText(p.PaidDate)
		}
	}
}
