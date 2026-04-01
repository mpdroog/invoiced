package invoice

import (
	"github.com/mpdroog/invoiced/model"
)

// InvoiceMail is an alias for model.InvoiceMail.
type InvoiceMail = model.InvoiceMail //nolint:revive // backwards compatibility

// InvoiceEntity is an alias for model.InvoiceEntity.
type InvoiceEntity = model.InvoiceEntity //nolint:revive // backwards compatibility

// InvoiceCustomer is an alias for model.InvoiceCustomer.
type InvoiceCustomer = model.InvoiceCustomer //nolint:revive // backwards compatibility

// InvoiceMeta is an alias for model.InvoiceMeta.
type InvoiceMeta = model.InvoiceMeta //nolint:revive // backwards compatibility

// InvoiceLine is an alias for model.InvoiceLine.
type InvoiceLine = model.InvoiceLine //nolint:revive // backwards compatibility

// InvoiceTotal is an alias for model.InvoiceTotal.
type InvoiceTotal = model.InvoiceTotal //nolint:revive // backwards compatibility

// InvoiceBank is an alias for model.InvoiceBank.
type InvoiceBank = model.InvoiceBank //nolint:revive // backwards compatibility

// Invoice is an alias for model.Invoice.
type Invoice = model.Invoice //nolint:revive // backwards compatibility

// ListReply contains the response for listing invoices.
type ListReply struct {
	Invoices map[string][]*Invoice
}
