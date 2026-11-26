// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package domain

import (
	"testing"

	"github.com/movaltech/decimal"
)

func TestTaxNominal(t *testing.T) {
	cases := []struct {
		name string
		tax  Tax
		want bool
	}{
		{"IVA", Tax{TypeCode: "01", Percent: decimal.MustParse("19")}, false},
		{"ICUI, a percentage", Tax{TypeCode: "35", Percent: decimal.MustParse("20")}, false},
		{"IBUA", Tax{TypeCode: "34", BaseUnitCode: "MLT", BaseUnitMeasure: decimal.MustParse("25000"), PerUnitAmountCents: 1800}, true},
		// The unit code alone decides: the builder renders what the caller set.
		{"unit code without amounts", Tax{TypeCode: "22", BaseUnitCode: "NIU"}, true},
	}
	for _, c := range cases {
		if got := c.tax.Nominal(); got != c.want {
			t.Errorf("%s: Nominal() = %v, want %v", c.name, got, c.want)
		}
	}
}
