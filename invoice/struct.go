package invoice

import (
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/model"
)

// Type aliases for backwards compatibility
type InvoiceMail = model.InvoiceMail
type InvoiceEntity = model.InvoiceEntity
type InvoiceCustomer = model.InvoiceCustomer
type InvoiceMeta = model.InvoiceMeta
type InvoiceLine = model.InvoiceLine
type InvoiceTotal = model.InvoiceTotal
type InvoiceBank = model.InvoiceBank
type Invoice = model.Invoice

type ListReply struct {
	Invoices map[string][]*Invoice
	Commits  []*db.CommitMessage
}
