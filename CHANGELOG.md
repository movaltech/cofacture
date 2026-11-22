# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this
project adheres to [Semantic Versioning](https://semver.org/).

## [2.5.0] — 2026-11-22

### Added

- `qr.InvoiceText` and `qr.NoteText`: the text of the QR printed on the graphic
  representation of an invoice and of a credit or debit note, as defined in section 11.7 of
  the invoicing annex 1.9: `NumFac`, `FecFac`, `HorFac`, `NitFac`, `DocAdq`, `ValFac`, `ValIva`,
  `ValOtroIm`, `ValTolFac`, `CUFE` (`CUDE` on the notes) and `QRCode`. `ValIva` is the IVA
  (tax 01), `ValOtroIm` every other tax; withholdings stay out. As with every document,
  `sts:QRCode` keeps the plain URL (`qr.URL`).

## [2.4.0] — 2026-11-22

### Added

- `qr.SupportDocumentText` and `qr.AdjustmentNoteText`: the text of the QR printed on the
  graphic representation of a support document and of its adjustment note, as defined in
  section 12.2 of the support document annex (Resolution 000167 of 2021): `NumDS`, `FecDS`,
  `HorDS`, `NumSNO`, `DocAdq`, `ValDS`, `ValIva`, `ValTolDS`, `CUDS` and `QRCode` (`NumNAS`,
  `FecNAS`, `HorNAS`, `ValNAS` and `ValTolNAS` on the note). `ValIva` adds up only the IVA
  (tax 01). Where the annex's example and its field definition differ (the example names the
  acquirer `NITABS`), the definition wins.

### Deprecated

- `qr.SupportDocumentContent` and `qr.AdjustmentNoteContent`. They put the software PIN,
  which is secret, in the QR text, used field names outside the annex and added `Amb`. They
  now return the new annex text and ignore `softwarePIN`. Neither block goes in `sts:QRCode`:
  annex rules DSAB36 and NSAB36 ask for the plain URL there (`qr.SupportDocumentURL`,
  `qr.URL`).

## [2.3.0] — 2026-11-22

### Added

- `domain.Line.BrandName` and `ModelName`: the item's brand and model, written as
  `cac:Item/cbc:BrandName` and `cbc:ModelName` after its description, in every document with
  lines. Both are optional; DIAN notifies their absence (FAZ04/FAZ05 on invoices, NSAZ04/NSAZ05
  on adjustment notes) and accepts the document, so they stay empty when the item has none.

## [2.2.2] — 2026-11-19

### Fixed

- `cac:TaxTotal` is one per tax type, with the sum of that type's subtotals, at header and
  line level. Before, every tax went under a single `TaxTotal`, which DIAN rejects with FAS01b
  and FAS02 when a document mixes tax types, e.g. IVA and INC; documents with a single tax
  type, such as IVA at several rates, build as before.
- `soap.NumberRangeResponseList` documents that DIAN answers `OperationCode` "100" ("Accion
  completada OK.") on success, besides "0", as habilitacion does, and that a support document
  range (prefix SEDS) comes without a technical key.

## [2.2.1] — 2026-11-19

### Fixed

- Release 2.2.0 was tagged on the 2.1.0 commit by mistake, so `v2.2.0` has exactly the content
  of 2.1.0 and none of the changes below. 2.2.1 is the release that carries them, and `go.mod`
  retracts `v2.2.0`, so `go get` skips it.

## [2.2.0] — 2026-11-19 [YANKED]

### Added

- Discounts and charges (`cac:AllowanceCharge`): `domain.AllowanceCharge` on a `Line` (it
  changes the line's taxable base, rule FAV06) and on the `Document` (it changes no base, rule
  FAQ01), with `Totals.AllowanceTotalCents` and `Totals.ChargeTotalCents`
  (`cbc:AllowanceTotalAmount`, `cbc:ChargeTotalAmount`). The Invoice, Credit Note and Debit Note
  builders render them where each schema puts them, which differs: on the Debit Note the
  document-level group goes before `cac:PaymentMeans`, and in Credit and Debit Note lines it
  goes after `cac:TaxTotal`. IDs are numbered from 1 (rule FAQ02).
- Advance payments: `domain.PrepaidPayment` and `Document.PrepaidPayments` render
  `cac:PrepaidPayment` (group FBD01) on the Invoice; their sum is still `Totals.PrepaidCents`.
- `Document.WithholdingTaxes` is rendered on the Invoice too (`cac:WithholdingTaxTotal`), where
  rule FAT01 limits it to the self-withholdings (autorretenciones) the issuer practices. It
  changes no total and carries no `cbc:TaxExemptionReason`, which only the Support Document and
  its Adjustment Note have. Credit and Debit Notes ignore it: their schemas have no such group.

Verified in DIAN habilitacion: invoices, credit notes and debit notes with line and document
discounts and charges, ReteFuente, ReteIVA and an advance payment were accepted without
notices. DIAN answers notice FAT07 for any ReteICA (07) withholding, which its table 13.3.11
gives no rates for. Documents that use none of these fields build exactly as in 2.1.0.

## [2.1.0] — 2026-11-19

### Added

- `signer.LoadPKCS12` reads BER-encoded `.p12`/`.pfx` files, as exported by Java tools and
  issued by several Colombian certificate authorities (indefinite lengths, content split in
  1000-byte chunks, legacy RC2-40 encryption). They load as they are; before, they failed with
  `asn1: syntax error: indefinite length found (not DER)` and had to be re-exported with OpenSSL.

### Changed

- PKCS#12 decoding moved from `software.sslmate.com/src/go-pkcs12` v0.7.2 to an internal copy
  (`internal/pkcs12`, BSD license kept in `internal/pkcs12/LICENSE`) that re-encodes every
  structure from BER to DER before parsing it. The MAC is still verified over the original
  bytes. DER files decode exactly as before, and the upstream test suite still passes.
  `golang.org/x/crypto` becomes a direct dependency.

## [2.0.1] — 2026-11-19

### Fixed

- The doc comment of `domain.Party.EntityTypeCode` had the codes backwards. DIAN table 13.2.3
  (TipoOrganizacion) and Technical Annex 1.9 rule FAK19 define `"1"` as legal entity (persona
  jurídica) and `"2"` as natural person (persona natural). The builder already behaved
  accordingly (it adds `cac:PartyIdentification` for `"2"`, as rule FAK61 expects for natural
  persons), but callers following the comment sent the codes swapped. The sample parties in the
  builder tests and their golden files now use the official codes. No API change.

## [2.0.0] — 2026-11-19

The module path is now `github.com/movaltech/cofacture/v2`. Code still using v1 keeps working
unchanged; to upgrade, run `go get github.com/movaltech/cofacture/v2` and replace
`github.com/movaltech/cofacture/` with `github.com/movaltech/cofacture/v2/` in your imports.

### Changed

- **Breaking:** every `soap.Client` method takes a `context.Context` as its first argument,
  e.g. `client.SendBillSync(ctx, fileName, zip)`. The request is built with
  `http.NewRequestWithContext`, so canceling the context or passing its deadline aborts an
  in-flight call; `HTTPClient.Timeout` still applies on top of it.
- **Breaking:** quantities and percentages use `decimal.Decimal`
  ([github.com/movaltech/decimal](https://github.com/movaltech/decimal)) instead of `float64`:
  `domain.Line.Quantity`, `domain.Tax.Percent`, `payroll.DeduccionPct.Porcentaje`,
  `payroll.FondoSP.Porcentaje`/`PorcentajeSub`, and `soap.TotalEImpuestos.Iva`/`Total`. Values are
  rendered exactly (round half away from zero), so `2.675` at 2 decimals is now `2.68`; float64
  with `%.2f` printed `2.67`. The XML generated for ordinary values is byte-identical to v1.
- **Breaking:** exported identifiers follow Go initialisms: `XmlBase64Bytes`, `XmlBytes`,
  `XmlBytesBase64`, `XmlDocumentKey`, `XmlFileName`, `DocumentTypeId`,
  `XMLParamsResponseTrackId` and `GetXmlByDocumentKey` are now `XMLBase64Bytes`, `XMLBytes`,
  `XMLBytesBase64`, `XMLDocumentKey`, `XMLFileName`, `DocumentTypeID`,
  `XMLParamsResponseTrackID` and `GetXMLByDocumentKey`. XML tags and SOAP action names on the
  wire are unchanged.
- SOAP responses are read with a 64 MiB limit.
- Requires Go 1.26.5 (the minimum of `github.com/movaltech/decimal`).

### Added

- `domain.FormatFixed(d, places)`, the exact fixed-decimal formatter used for quantities and
  percentages.

### Fixed

- A `.gitattributes` file forces LF line endings. With `core.autocrlf=true` on Windows, the
  golden-file tests failed after checkout because the XML fixtures were converted to CRLF.

## [1.0.0] — 2026-09-14

### Changed

- Module path is now `github.com/movaltech/cofacture`. No API changes from 0.2.0.

## [0.2.0] — 2026-09-04

### Changed

- **Breaking:** renamed `domain.Invoice` to `domain.Document` (file `domain/invoice.go` →
  `domain/document.go`). It was never Invoice-only — the same struct is the model for Support
  Document and Documento Equivalente Electrónico, and is embedded in `CreditNote`, `DebitNote`,
  and `AdjustmentNote`. The new name and doc comment reflect that; callers using `domain.Invoice`
  or the `.Invoice` embedded-field selector need to update to `domain.Document`/`.Document`.

## [0.1.2] — 2026-09-04

### Fixed

- Corrected a stale function-name reference in a doc comment (`appendSupplierParty` →
  `appendDSSupplierParty`).

### Testing

- Added regression tests for `signer.LoadPKCS12`'s CA-chain fix, and for `LoadPEM` — both were
  previously untested.
- Added tests for `qr.SupportDocumentURL`, `qr.AdjustmentNoteContent`, and the tax-fallback
  branches of both QR content builders (coverage 48.8% → 95.1%).
- Added tests for `internal/dianhash.Seed`, previously untested (coverage 0% → 100%).

## [0.1.1] — 2026-09-04

### Fixed

- Corrected the expected `ProfileID` value for Invoice, Credit Note, and Debit Note — found
  while testing against DIAN's certification environment. The generic `"DIAN 2.1"` does not
  match what the Technical Annex expects in `cbc:ProfileID`; it must name the specific document,
  e.g. `"DIAN 2.1: Factura Electrónica de Venta"`, `"DIAN 2.1: Nota Crédito de Factura
  Electrónica de Venta"`, or `"DIAN 2.1: Nota Débito de Factura Electrónica de Venta"`.
  `ProfileID` is a caller-supplied field on `domain.Invoice`, not hardcoded by the builder, so
  callers who used the old `"DIAN 2.1"` value should update it.
- `signer.LoadPKCS12` now accepts `.p12`/`.pfx` files that bundle a CA chain alongside the leaf
  certificate — the previous implementation (`pkcs12.Decode`) rejected any file with more than
  the certificate and key ("expected exactly two safe bags"), which real DIAN-issued certificates
  routinely violate. Switched to `pkcs12.DecodeChain`; the chain itself is discarded, since DIAN's
  XAdES-EPES signature only ever embeds the leaf certificate.

### Documentation

- Documented a DIAN business rule confirmed against real submissions: the issuer's identification
  must use `TypeCode: "31"` (NIT) — not `"13"` (cédula) — for Debit Notes, and the acquirer's must
  do the same for Support Documents, even when the same party identifies with `"13"` without
  issue on an Invoice or Credit Note. Also documented that `TaxSchemeCode`/`TaxSchemeName` must
  reflect a real tax regime once a party is NIT-identified. Doc comments only, no behavior change.
- Added a "Document coverage" table to the README showing certification status per document
  type, and updated the Quick Start example to fetch the numbering range via
  `GetNumberingRange`. General wording cleanup throughout.

## [0.1.0] — 2026-09-04

Initial release.

### Added

- UBL 2.1 document builders: Electronic Sales Invoice, Credit Note, Debit Note, Support
  Document, Adjustment Note to the Support Document, Attached Document, and Documento
  Equivalente Electrónico (POS ticket and other sub-types).
- Individual Electronic Payroll builder (`NominaIndividual`) and its adjustment/cancellation
  variants.
- RADIAN event builders: Acuse de Recibo, Reclamo, Recibo del Bien, Aceptación Expresa,
  Aceptación Tácita.
- CUFE, CUDE, CUDS, and CUNE computation per the DIAN Technical Annexes.
- XAdES-EPES signing (C14N 1.0, RSA-SHA256), loading certificates from PEM or PKCS#12.
- ZIP packaging matching the file-naming convention DIAN's receiving service requires.
- SOAP 1.2 + WS-Security client implementing all 16 operations of the `WcfDianCustomerServices`
  contract, for both habilitación and producción.
- Response parser turning DIAN's validation output into a structured result.

[0.2.0]: https://github.com/movaltech/cofacture/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/movaltech/cofacture/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/movaltech/cofacture/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/movaltech/cofacture/releases/tag/v0.1.0
