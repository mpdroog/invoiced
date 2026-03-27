package purchase

// PurchaseInvoice represents an incoming invoice from a supplier
type PurchaseInvoice struct {
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

type Supplier struct {
	Name  string
	VAT   string
	COC   string
	Email string
}

type PurchaseLine struct {
	Description string
	Quantity    string
	Price       string
	Total       string
	TaxPercent  string
}

type ListReply struct {
	Invoices map[string][]*PurchaseInvoice
}
