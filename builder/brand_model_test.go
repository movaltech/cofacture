// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"strings"
	"testing"

	"github.com/movaltech/cofacture/v2/domain"
)

// TestItemBrandAndModel writes an item's brand and model after its description, in the UBL
// 2.1 order, and leaves them out when the line has none.
func TestItemBrandAndModel(t *testing.T) {
	inv := sampleInvoice()
	inv.Lines[0].BrandName, inv.Lines[0].ModelName = "Acme", "X-200"
	for name, build := range map[string]func(domain.Document) (string, error){
		"invoice": func(d domain.Document) (string, error) {
			doc, err := BuildInvoice(d)
			if err != nil {
				return "", err
			}
			return doc.WriteToString()
		},
		"support document": func(d domain.Document) (string, error) {
			d.DocumentTypeCode, d.HashType = "05", "CUDS-SHA384"
			doc, err := BuildSupportDocument(d)
			if err != nil {
				return "", err
			}
			return doc.WriteToString()
		},
	} {
		xml, err := build(inv)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		want := "<cbc:Description>" + inv.Lines[0].Description + "</cbc:Description><cbc:BrandName>Acme</cbc:BrandName><cbc:ModelName>X-200</cbc:ModelName>"
		if !strings.Contains(xml, want) {
			t.Errorf("%s: item does not carry brand and model after its description", name)
		}
	}

	plain := sampleInvoice()
	doc, err := BuildInvoice(plain)
	if err != nil {
		t.Fatal(err)
	}
	xml, _ := doc.WriteToString()
	if strings.Contains(xml, "BrandName") || strings.Contains(xml, "ModelName") {
		t.Error("a line without brand and model must not carry the elements")
	}
}
