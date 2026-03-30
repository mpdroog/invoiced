package model

// InvoiceMail contains email settings for an invoice.
type InvoiceMail struct {
	From    string
	Subject string
	To      string
	Body    string
}

// InvoiceEntity contains the seller/entity details.
type InvoiceEntity struct {
	Name    string `validate:"nonzero"`
	Street1 string `validate:"nonzero"`
	Street2 string `validate:"nonzero"`
}

// InvoiceCustomer contains the buyer/customer details.
type InvoiceCustomer struct {
	Name    string `validate:"nonzero"`
	Street1 string `validate:"nonzero"`
	Street2 string `validate:"nonzero"`
	Vat     string
	Coc     string
	Tax     string // Simple string so we know what to tax
}

// InvoiceMeta contains invoice metadata like dates and IDs.
type InvoiceMeta struct {
	Conceptid string `validate:"slug"`
	Status    string `validate:"slug"`
	Invoiceid string `validate:"slug"`
	Issuedate string `validate:"date"`
	Ponumber  string `validate:"slug"`
	Duedate   string `validate:"nonzero,date"`
	Paydate   string `validate:"date"`
	Freefield string
	HourFile  string
}

// InvoiceLine represents a single line item on an invoice.
type InvoiceLine struct {
	Description string `validate:"nonzero"`
	Quantity    string `validate:"nonzero,qty"`
	Price       string `validate:"nonzero,price"`
	Total       string `validate:"nonzero,price"`
}

// InvoiceTotal contains the invoice totals.
type InvoiceTotal struct {
	Ex    string `validate:"nonzero,price"`
	Tax   string `validate:"nonzero,price"`
	Total string `validate:"nonzero,price"`
}

// InvoiceBank contains banking details for payment.
type InvoiceBank struct {
	Vat  string `validate:"nonzero"`
	Coc  string `validate:"nonzero"`
	Iban string `validate:"nonzero,iban"`
	Bic  string `validate:"nonzero,bic"`
}

// Invoice represents a complete invoice document.
type Invoice struct {
	Company  string `validate:"nonzero"`
	Entity   InvoiceEntity
	Customer InvoiceCustomer
	Meta     InvoiceMeta
	Lines    []InvoiceLine
	Notes    string
	Total    InvoiceTotal
	Bank     InvoiceBank
	Mail     InvoiceMail
}
