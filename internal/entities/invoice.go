package entities

// InvoiceDocumentType selects which of the two documents Generate renders
// for a Sale — see InvoiceGenerate.DocumentType.
type InvoiceDocumentType int32

const (
	InvoiceDocumentTypeUnspecified InvoiceDocumentType = iota
	// InvoiceDocumentTypeOffer is the pre-payment offer/quote (Serbian:
	// "ponuda") — available for a Sale in any status, matching the
	// document this service rendered before DocumentType existed.
	InvoiceDocumentTypeOffer
	// InvoiceDocumentTypeReceipt is the post-payment receipt (Serbian:
	// "račun") — Generate rejects it (ErrInvoiceReceiptRequiresPayment)
	// while the Sale is still SaleStatusDraft, since nothing has been
	// paid yet to receipt.
	InvoiceDocumentTypeReceipt
)

// InvoiceGenerate is InvoicesService.Generate's input — renders an existing
// Sale as a PDF (see DocumentType). Stateless: nothing about the rendered
// document is persisted against the Sale, only the buyer legal-field
// overrides (which are merged onto the underlying Client/Partner so they
// don't have to be retyped next time). Seller letterhead, commercial
// terms, and VAT rate all come from CRMConfig.Invoice, not from this
// request.
type InvoiceGenerate struct {
	SaleID string `normalize:"trim"`
	// DocumentType picks Offer (any Sale status) or Receipt (requires a
	// paid-or-later Sale status) — defaults to InvoiceDocumentTypeOffer
	// when left unspecified.
	DocumentType InvoiceDocumentType
	// Buyer overrides — when non-empty, merged onto the Sale's Client (or
	// Partner, for a partner-only sale) via that entity's own Update, so
	// they're remembered for next time instead of being retyped per
	// invoice.
	BuyerTaxID              string  `normalize:"trim"`
	BuyerRegistrationNumber string  `normalize:"trim"`
	BuyerCode               string  `normalize:"trim"`
	BuyerAddress            *string `normalize:"trim,nil_on_empty"`
	// IncludeVAT selects whether each line/total shows VAT at
	// CRMConfig.Invoice.VAT.Percentage, or omits the VAT column/rows
	// entirely.
	IncludeVAT bool
	// SendEmail, when true, additionally emails the rendered PDF to the
	// buyer (EmailOverride, or the Client/Partner's own email) — best
	// effort: a delivery failure is logged, not returned as an error, since
	// the PDF itself was already generated successfully.
	SendEmail     bool
	EmailOverride *string `normalize:"trim,nil_on_empty"`
}
