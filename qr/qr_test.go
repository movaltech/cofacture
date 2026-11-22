// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package qr

import (
	"strings"
	"testing"

	"github.com/movaltech/cofacture/v2/domain"
)

func TestURL(t *testing.T) {
	cases := []struct {
		env, want string
	}{
		{"1", "https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey=abc123"},
		{"2", "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=abc123"},
	}
	for _, c := range cases {
		if got := URL(c.env, "abc123"); got != c.want {
			t.Errorf("URL(%q, ...) = %q, want %q", c.env, got, c.want)
		}
	}
}

func TestSupportDocumentURL(t *testing.T) {
	cases := []struct {
		env, want string
	}{
		{"1", "https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey=cuds123"},
		{"2", "https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=cuds123"},
	}
	for _, c := range cases {
		if got := SupportDocumentURL(c.env, "cuds123"); got != c.want {
			t.Errorf("SupportDocumentURL(%q, ...) = %q, want %q", c.env, got, c.want)
		}
	}
}

// sample is a support document with IVA at two rates and an INC, so ValIva must add up only
// the IVA (tax 01).
func sample(env string) domain.Document {
	return domain.Document{
		Prefix: "SEDS", Number: "984000014", IssueDate: "2026-11-20", IssueTime: "09:30:00-05:00",
		EnvironmentCode: env,
		HeaderTaxes: []domain.Tax{
			{TypeCode: "01", TaxAmountCents: 1_900_000},
			{TypeCode: "01", TaxAmountCents: 50_000},
			{TypeCode: "04", TaxAmountCents: 80_000},
		},
		Totals: domain.Totals{LineExtensionCents: 11_000_000, PayableCents: 13_030_000},
		// Support document roles: Supplier is the seller (SNO), Customer the acquirer.
		Supplier: domain.Party{Identification: domain.Identification{Number: "1020304050"}},
		Customer: domain.Party{Identification: domain.Identification{Number: "900123456"}},
	}
}

const cuds = "907e4444decc9e59c160a2fb3b6659b33dc5b632a5008922b9a62f83f757b1c448e47f5867f2b50dbdb96f48c7681168"

// TestSupportDocumentText checks the block of annex 12.2.1, field by field and in order.
func TestSupportDocumentText(t *testing.T) {
	want := "NumDS: SEDS984000014\n" +
		"FecDS: 2026-11-20\n" +
		"HorDS: 09:30:00-05:00\n" +
		"NumSNO: 1020304050\n" +
		"DocAdq: 900123456\n" +
		"ValDS: 110000.00\n" +
		"ValIva: 19500.00\n" +
		"ValTolDS: 130300.00\n" +
		"CUDS: " + cuds + "\n" +
		"QRCode: https://catalogo-vpfe-hab.dian.gov.co/document/searchqr?documentkey=" + cuds
	if got := SupportDocumentText(sample("2"), cuds); got != want {
		t.Errorf("SupportDocumentText() =\n%s\nwant\n%s", got, want)
	}
}

// TestAdjustmentNoteText checks the block of annex 12.2.2, which names the note's fields NAS.
func TestAdjustmentNoteText(t *testing.T) {
	doc := sample("1")
	doc.Prefix, doc.Number = "NA", "1"
	want := "NumNAS: NA1\n" +
		"FecNAS: 2026-11-20\n" +
		"HorNAS: 09:30:00-05:00\n" +
		"NumSNO: 1020304050\n" +
		"DocAdq: 900123456\n" +
		"ValNAS: 110000.00\n" +
		"ValIva: 19500.00\n" +
		"ValTolNAS: 130300.00\n" +
		"CUDS: " + cuds + "\n" +
		"QRCode: https://catalogo-vpfe.dian.gov.co/document/searchqr?documentkey=" + cuds
	if got := AdjustmentNoteText(doc, cuds); got != want {
		t.Errorf("AdjustmentNoteText() =\n%s\nwant\n%s", got, want)
	}
}

// TestTextWithoutVAT writes ValIva as zero when the document has no IVA.
func TestTextWithoutVAT(t *testing.T) {
	doc := sample("2")
	doc.HeaderTaxes = []domain.Tax{{TypeCode: "04", TaxAmountCents: 80_000}}
	if got := SupportDocumentText(doc, cuds); !strings.Contains(got, "\nValIva: 0.00\n") {
		t.Errorf("want ValIva: 0.00, got:\n%s", got)
	}
}

// TestDeprecatedContentHasNoPIN: the old functions now return the annex text and drop the PIN.
func TestDeprecatedContentHasNoPIN(t *testing.T) {
	doc := sample("2")
	for name, got := range map[string]string{
		"SupportDocumentContent": SupportDocumentContent(doc, cuds, "40125"),
		"AdjustmentNoteContent":  AdjustmentNoteContent(doc, cuds, "40125"),
	} {
		if strings.Contains(got, "40125") || strings.Contains(got, "PIN") {
			t.Errorf("%s carries the software PIN:\n%s", name, got)
		}
	}
	if SupportDocumentContent(doc, cuds, "x") != SupportDocumentText(doc, cuds) ||
		AdjustmentNoteContent(doc, cuds, "x") != AdjustmentNoteText(doc, cuds) {
		t.Error("the deprecated functions must return the annex text")
	}
}
