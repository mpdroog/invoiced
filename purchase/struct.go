package purchase

// PurchaseInvoice represents an incoming invoice from a supplier.
type PurchaseInvoice struct { //nolint:revive // maintaining public API
	ID          string // Original invoice ID from supplier
	Supplier    Supplier
	Issuedate   string
	Duedate     string
	TotalEx     string
	TotalTax    string
	TotalInc    string
	Currency    string
	PaymentRef  string // Payment reference (e.g. RF number)
	IBAN        string
	BIC         string
	Lines       []PurchaseLine
	PDFFilename string // Embedded PDF filename
	XMLFilename string // Original XML filename
	Status      string // UNPAID, PAID
	Paydate     string
}

// Supplier contains information about the invoice supplier.
type Supplier struct {
	Name  string
	VAT   string
	COC   string
	Email string
}

// PurchaseLine represents a single line item on a purchase invoice.
type PurchaseLine struct { //nolint:revive // maintaining public API
	Description string
	Quantity    string
	Price       string
	Total       string
	TaxPercent  string
}

// ListReply contains the response for listing purchase invoices.
type ListReply struct {
	Invoices map[string][]*PurchaseInvoice
}
