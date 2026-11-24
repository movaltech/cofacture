// Copyright (c) 2026 Diego Montoya
// SPDX-License-Identifier: AGPL-3.0

package builder

import (
	"errors"

	"github.com/beevik/etree"
	"github.com/movaltech/cofacture/v2/domain"
	ubl "github.com/movaltech/cofacture/v2/xml"
)

// SignaturePlaceholder locates the empty ext:ExtensionContent reserved for the XAdES signature
// (the last ext:UBLExtension of ext:UBLExtensions) in a document already built by BuildInvoice
// or an equivalent. The signer package uses it as the insertion point.
func SignaturePlaceholder(doc *etree.Document) (*etree.Element, error) {
	extensions := doc.FindElement("//ext:UBLExtensions")
	if extensions == nil {
		return nil, errors.New("builder: the document has no ext:UBLExtensions")
	}
	children := extensions.SelectElements("ext:UBLExtension")
	if len(children) == 0 {
		return nil, errors.New("builder: ext:UBLExtensions has no ext:UBLExtension")
	}
	last := children[len(children)-1]
	content := last.SelectElement("ext:ExtensionContent")
	if content == nil {
		return nil, errors.New("builder: the last ext:UBLExtension has no ext:ExtensionContent")
	}
	if len(content.ChildElements()) > 0 {
		return nil, errors.New("builder: the signature placeholder already has content")
	}
	return content, nil
}

// invoiceControlTypes are the document types whose sts:DianExtensions carry sts:InvoiceControl
// (the numbering authorized by DIAN): the Invoice ("01"), the Support Document ("05") and the
// POS ticket ("20"; Technical Annex Documento Equivalente Electrónico V1.0, DEAB04-DEAB13,
// "only for the POS"), also when it goes out in contingency ("07" by the issuer, "08" by DIAN;
// annex section 15, same prefix and number) with its numbering range set.
var invoiceControlTypes = map[string]bool{"01": true, "05": true, "20": true}

// appendUBLExtensions adds ext:UBLExtensions with the DIAN extensions (sts:DianExtensions), the
// Documento Equivalente Electrónico extensions the document carries (software manufacturer,
// buyer benefits, cash register), and a last, empty ext:UBLExtension, reserved for the XAdES
// signature that the signer package adds in a later pipeline step.
func appendUBLExtensions(parent *etree.Element, inv domain.Document) {
	extensions := parent.CreateElement("ext:UBLExtensions")

	dianExt := extensions.CreateElement("ext:UBLExtension").
		CreateElement("ext:ExtensionContent").
		CreateElement("sts:DianExtensions")

	contingency := inv.DocumentTypeCode == "07" || inv.DocumentTypeCode == "08"
	if invoiceControlTypes[inv.DocumentTypeCode] || (contingency && inv.NumberingRange.AuthorizedCode != "") {
		appendInvoiceControl(dianExt, inv.NumberingRange)
	}

	source := dianExt.CreateElement("sts:InvoiceSource").CreateElement("cbc:IdentificationCode")
	source.CreateAttr("listAgencyID", "6")
	source.CreateAttr("listAgencyName", "United Nations Economic Commission for Europe")
	source.CreateAttr("listSchemeURI", "urn:oasis:names:specification:ubl:codelist:gc:CountryIdentificationCode-2.1")
	source.SetText("CO")

	swProvider := dianExt.CreateElement("sts:SoftwareProvider")
	providerID := swProvider.CreateElement("sts:ProviderID")
	providerID.CreateAttr("schemeAgencyID", ubl.DianSchemeAgencyID)
	providerID.CreateAttr("schemeAgencyName", ubl.DianSchemeAgencyName)
	if inv.SoftwareProvider.ProviderIdentification.TypeCode == ubl.IdentificationTypeNIT {
		providerID.CreateAttr("schemeID", inv.SoftwareProvider.ProviderIdentification.VerificationCode)
	}
	providerID.CreateAttr("schemeName", inv.SoftwareProvider.ProviderIdentification.TypeCode)
	providerID.SetText(inv.SoftwareProvider.ProviderIdentification.Number)

	softwareID := swProvider.CreateElement("sts:SoftwareID")
	softwareID.CreateAttr("schemeAgencyID", ubl.DianSchemeAgencyID)
	softwareID.CreateAttr("schemeAgencyName", ubl.DianSchemeAgencyName)
	softwareID.SetText(inv.SoftwareProvider.SoftwareID)

	secCode := dianExt.CreateElement("sts:SoftwareSecurityCode")
	secCode.CreateAttr("schemeAgencyID", ubl.DianSchemeAgencyID)
	secCode.CreateAttr("schemeAgencyName", ubl.DianSchemeAgencyName)
	secCode.SetText(inv.SoftwareSecurityCode)

	authProviderID := dianExt.CreateElement("sts:AuthorizationProvider").CreateElement("sts:AuthorizationProviderID")
	authProviderID.CreateAttr("schemeAgencyID", ubl.DianSchemeAgencyID)
	authProviderID.CreateAttr("schemeAgencyName", ubl.DianSchemeAgencyName)
	authProviderID.CreateAttr("schemeID", "4")
	authProviderID.CreateAttr("schemeName", "31")
	authProviderID.SetText(ubl.DianAuthorizationProviderID)

	dianExt.CreateElement("sts:QRCode").SetText(inv.QRURL)

	appendEquivalentExtensions(extensions, inv)

	// Reserved for the XAdES signature (signer.Sign fills it in later).
	extensions.CreateElement("ext:UBLExtension").CreateElement("ext:ExtensionContent")
}

func appendInvoiceControl(parent *etree.Element, nr domain.NumberingRange) {
	control := parent.CreateElement("sts:InvoiceControl")
	control.CreateElement("sts:InvoiceAuthorization").SetText(nr.AuthorizedCode)

	period := control.CreateElement("sts:AuthorizationPeriod")
	period.CreateElement("cbc:StartDate").SetText(nr.StartDate)
	period.CreateElement("cbc:EndDate").SetText(nr.EndDate)

	authorized := control.CreateElement("sts:AuthorizedInvoices")
	if nr.Prefix != "" {
		authorized.CreateElement("sts:Prefix").SetText(nr.Prefix)
	}
	authorized.CreateElement("sts:From").SetText(nr.StartNumber)
	authorized.CreateElement("sts:To").SetText(nr.EndNumber)
}

// appendEquivalentExtensions adds the Documento Equivalente Electrónico extensions that are set,
// each in its own ext:UBLExtension, as Name/Value pairs (Technical Annex Documento Equivalente
// Electrónico V1.0): FabricanteSoftware (section 8.1), BeneficiosComprador (8.2.1.1; the annex tables
// say "Veneficios", its rule DEPD15 and the published examples say "Beneficios") and PuntoVenta
// (8.2.1.2). The Name literals are the annex's, accents included ("UbicaciónCaja", "CódigoVenta").
func appendEquivalentExtensions(extensions *etree.Element, inv domain.Document) {
	if m := inv.SoftwareManufacturer; m != nil {
		pairs := [][2]string{{"NombreApellido", m.Name}}
		if m.LegalName != "" {
			pairs = append(pairs, [2]string{"RazonSocial", m.LegalName})
		}
		pairs = append(pairs, [2]string{"NombreSoftware", m.SoftwareName})
		appendNameValueExtension(extensions, "FabricanteSoftware", "InformacionDelFabricanteDelSoftware", pairs)
	}
	if b := inv.POSBuyerBenefits; b != nil {
		appendNameValueExtension(extensions, "BeneficiosComprador", "InformacionBeneficiosComprador", [][2]string{
			{"Codigo", b.Code}, {"NombresApellidos", b.Name}, {"Puntos", b.Points},
		})
	}
	if c := inv.POSCashRegister; c != nil {
		appendNameValueExtension(extensions, "PuntoVenta", "InformacionCajaVenta", [][2]string{
			{"PlacaCaja", c.Plate}, {"UbicaciónCaja", c.Location}, {"Cajero", c.Cashier},
			{"TipoCaja", c.Type}, {"CódigoVenta", c.SaleCode}, {"SubTotal", formatAmount(c.SubtotalCents)},
		})
	}
}

// appendNameValueExtension adds ext:UBLExtension/ext:ExtensionContent/<group>/<info> with the
// pairs as consecutive Name and Value elements.
func appendNameValueExtension(extensions *etree.Element, group, info string, pairs [][2]string) {
	el := extensions.CreateElement("ext:UBLExtension").
		CreateElement("ext:ExtensionContent").
		CreateElement(group).
		CreateElement(info)
	for _, p := range pairs {
		el.CreateElement("Name").SetText(p[0])
		el.CreateElement("Value").SetText(p[1])
	}
}
