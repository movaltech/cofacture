# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this
project adheres to [Semantic Versioning](https://semver.org/).

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
