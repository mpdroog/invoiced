package invoice

import "github.com/mpdroog/invoiced/db"

type InvoiceMail struct {
	From    string
	Subject string
	To      string
	Body    string
}
type InvoiceEntity struct {
	Name    string `validate:"nonzero"`
	Street1 string `validate:"nonzero"`
	Street2 string `validate:"nonzero"`
}
type InvoiceCustomer struct {
	Name    string `validate:"nonzero"`
	Street1 string `validate:"nonzero"`
	Street2 string `validate:"nonzero"`
	Vat     string
	Coc     string
	Tax     string // Simple string so we know what to tax
}
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
type InvoiceLine struct {
	Description string `validate:"nonzero"`
	Quantity    string `validate:"nonzero,qty"`
	Price       string `validate:"nonzero,price"`
	Total       string `validate:"nonzero,price"`
}
type InvoiceTotal struct {
	Ex    string `validate:"nonzero,price"`
	Tax   string `validate:"nonzero,price"`
	Total string `validate:"nonzero,price"`
}
type InvoiceBank struct {
	Vat  string `validate:"nonzero"`
	Coc  string `validate:"nonzero"`
	Iban string `validate:"nonzero,iban"`
	Bic  string `validate:"nonzero,bic"`
}

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

type ListReply struct {
	Invoices map[string][]*Invoice
	Commits  []*db.CommitMessage
}
