package invoice

import (
	"github.com/mpdroog/invoiced/db"
	"github.com/mpdroog/invoiced/entities"
	"github.com/mpdroog/invoiced/middleware"
	"time"
	"fmt"
	"github.com/shopspring/decimal"
	"strconv"
)

// Get start of day
func today() time.Time {
	txt := time.Now().Format("2006-01-02")
	t, e := time.Parse("2006-01-02", txt)
	if e != nil {
		panic(e)
	}
	return t
}

// Convert hours to concept invoice
func HourToInvoice(entity, year, project, name, hourStr, email string, t *db.Txn) error {
	prj, e := entities.GetProject(t, entity, project)
	if e != nil {
		return e
	}
	if prj == nil {
		return fmt.Errorf("No such project %s", project)
	}
	debtor, e := entities.GetDebtor(t, entity, prj.Debtor)
	if e != nil {
		return e
	}
	if debtor == nil {
		return fmt.Errorf("No such debtor %s", prj.Debtor)
	}

	company := middleware.CompanyByName(entity)
	if company == nil {
		return fmt.Errorf("No such company %s", entity)		
	}
	user := middleware.UserByEmail(email)
	if user == nil {
		return fmt.Errorf("No such email %s", email)		
	}

	d, e := time.ParseDuration(fmt.Sprintf("%dh", 24 * prj.DueDays))
	if e != nil {
		return e
	}

	hours, e := decimal.NewFromString(hourStr)
	if e != nil {
		return e
	}
	extotal := decimal.NewFromFloat(prj.HourRate).Mul(hours)

	tax := decimal.NewFromFloat(0)
	if debtor.TAX == "NL21" {
		tax = extotal.Div(decimal.NewFromFloat(100)).Mul(decimal.NewFromFloat(21))
	}
	total := extotal.Add(tax)

	today := today()
	c := &Invoice{
		Company: company.Name,
		Entity: InvoiceEntity{
			Name: user.Name,
			Street1: user.Address1,
			Street2: user.Address2,
		},
		Customer: InvoiceCustomer{
			Name: debtor.Name,
			Street1: debtor.Street1,
			Street2: debtor.Street2,
			Vat: debtor.VAT,
			Coc: debtor.COC,
		},
		Meta: InvoiceMeta{
			Conceptid: fmt.Sprintf("CONCEPT-%s", randStringBytesRmndr(6)),
			Status: "CONCEPT",
			Invoiceid: "",
			Issuedate: today.Format("2006-01-02"),
			Ponumber: prj.PO,
			Duedate: today.Add(d).Format("2006-01-02"),
		},
		Lines: []InvoiceLine{InvoiceLine{
			Description: name,
			Quantity: hourStr,
			Price: strconv.FormatFloat(prj.HourRate, 'f', -1, 64),
			Total: extotal.StringFixed(2),
		}},
		Notes: prj.NoteAdd,
		Total: InvoiceTotal{
			Ex: extotal.StringFixed(2),
			Tax: tax.StringFixed(2),
			Total: total.StringFixed(2),
		},
		Bank: InvoiceBank{
			Vat: company.VAT,
			Coc: company.COC,
			Iban: company.IBAN,
		},
	}

	path := fmt.Sprintf("%s/%s/concepts/sales-invoices/%s.toml", entity, year, c.Meta.Conceptid)
	return t.Save(path, c)
}