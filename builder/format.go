// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"github.com/movaltech/cofacture/v2/domain"
	"github.com/movaltech/decimal"
)

// formatAmount renders cents as the 2-decimal string DIAN requires.
func formatAmount(cents int64) string {
	return domain.FormatCents(cents)
}

// formatQuantity renders a quantity with 6 decimals, as the technical annex requires.
func formatQuantity(q decimal.Decimal) string {
	return domain.FormatFixed(q, 6)
}

// formatPercent renders a tax percentage with 2 decimals.
func formatPercent(p decimal.Decimal) string {
	return domain.FormatFixed(p, 2)
}
