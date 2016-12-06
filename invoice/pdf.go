package invoice

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
)

/**
  company: "RootDev",
  entity: {
    name: "M.P. Droog",
    street1: "Dorpsstraat 236a",
    street2: "Obdam, 1713HP, NL"
  },
  customer: {
    name: "XSNews B.V.",
    street1: "New Yorkstraat 9-13",
    street2: "1175 RD Lijnden"
  },
  meta: {
    invoiceid: "",
    issuedate: "issuedate",
    ponumber: "P/O",
    duedate: "due"
  },
  lines: [{
    description: "description",
    quantity: 1,
    price: "12.00",
    total: "12.00"
  }],
  notes: "",
  total: {
    ex: "200",
    tax: "1000",
    total: "1200"
  },
  bank: {
    vat: "VAT",
    coc: "COC",
    iban: "IBEN"
*/
func pdf(c *Invoice) (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(162, 162, 162)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
	pdf.AddPage()

	var lastY float64 = 15

	// Company name
	{
		pdf.SetXY(20, 15)
		pdf.SetFont("Helvetica", "B", 18)
		pdf.Cell(10, 30, c.Company)
	}

	// From
	{
		pdf.SetXY(130, 15)
		pdf.SetFont("Helvetica", "", 10)
		pdf.Cell(10, 30, "From")

		pdf.SetXY(143, 15)
		pdf.Cell(10, 30, c.Entity.Name)

		pdf.SetXY(143, 20)
		pdf.Cell(10, 30, c.Entity.Street1)

		pdf.SetXY(143, 25)
		pdf.Cell(10, 30, c.Entity.Street2)
	}

	// Invoice for
	{
		pdf.SetXY(20, 50)
		pdf.SetFont("Helvetica", "", 10)
		pdf.Cell(10, 30, "Invoice For")

		pdf.SetXY(40, 50)
		pdf.Cell(10, 30, c.Customer.Name)

		pdf.SetXY(40, 55)
		pdf.Cell(10, 30, c.Customer.Street1)

		pdf.SetXY(40, 60)
		pdf.Cell(10, 30, c.Customer.Street2)

		last := 65.0
		if len(c.Customer.Vat) > 0 {
			pdf.SetXY(40, last)
			pdf.Cell(10, 30, c.Customer.Vat)
			last = last +5
		}
		if len(c.Customer.Coc) > 0 {
			pdf.SetXY(40, last)
			pdf.Cell(10, 30, c.Customer.Coc)
		}
	}

	// Meta
	{
		lastY = 50
		pdf.SetFont("Helvetica", "", 10)

		pdf.SetXY(120, 50)
		pdf.Cell(10, 30, "Invoice ID")
		pdf.SetXY(143, 50)
		pdf.Cell(10, 30, c.Meta.Invoiceid)

		pdf.SetXY(120, 55)
		pdf.Cell(10, 30, "Issue Date")
		pdf.SetXY(143, 55)
		pdf.Cell(10, 30, c.Meta.Issuedate)

		pdf.SetXY(120, 60)
		pdf.Cell(10, 30, "P/O Number")
		pdf.SetXY(143, 60)
		pdf.Cell(10, 30, c.Meta.Ponumber)

		pdf.SetXY(120, 65)
		pdf.Cell(10, 30, "Duedate")
		pdf.SetXY(143, 65)
		pdf.Cell(10, 30, c.Meta.Duedate)

		/*for key, val := range c.Meta {
			pdf.SetXY(120, lastY)
			pdf.Cell(10, 30, key)

			pdf.SetXY(143, lastY)
			pdf.Cell(10, 30, val)
			lastY += 5
		}*/
	}

	// Header
	{
		pdf.SetXY(20, 80)
		pdf.SetFont("Helvetica", "", 10)
		pdf.Cell(10, 30, "Description")

		pdf.SetXY(126, 80)
		pdf.CellFormat(10, 30, "Qty", "", 1, "R", false, 0, "")

		pdf.SetXY(150, 80)
		pdf.CellFormat(20, 30, "Unit Price", "", 1, "R", false, 0, "")

		pdf.SetXY(170, 80)
		pdf.CellFormat(30, 30, "Amount", "", 1, "R", false, 0, "")

		pdf.Line(20, 97, 200, 97)
	}

	// Lines
	{
		lastY = 85
		for _, line := range c.Lines {
			pdf.SetXY(20, lastY)
			pdf.SetFont("Helvetica", "", 10)
			pdf.Cell(10, 30, line.Description)

			pdf.SetXY(126, lastY)
			pdf.CellFormat(10, 30, line.Quantity, "", 1, "R", false, 0, "")

			pdf.SetXY(150, lastY)
			pdf.CellFormat(20, 30, line.Price, "", 1, "R", false, 0, "")

			pdf.SetXY(170, lastY)
			pdf.CellFormat(30, 30, line.Total, "", 1, "R", false, 0, "")

			lastY += 5
		}
	}

	// Totals
	{
		pdf.SetXY(126, 100)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.Cell(10, 30, "Subtotal")

		pdf.SetXY(170, 100)
		pdf.CellFormat(30, 30, c.Total.Ex, "", 1, "R", false, 0, "")

		pdf.SetXY(126, 105)
		pdf.Cell(10, 30, "TAX (21%)")

		pdf.SetXY(170, 105)
		pdf.CellFormat(30, 30, c.Total.Tax, "", 1, "R", false, 0, "")

		pdf.SetXY(126, 120)
		pdf.SetFont("Helvetica", "B", 14)
		pdf.Cell(10, 30, "Amount Due")

		pdf.SetXY(170, 120)
		pdf.CellFormat(30, 30, string([]byte{byte(128)})+c.Total.Total, "0", 1, "R", false, 0, "")
	}

	// Notes
	{
		pdf.Line(20, 170, 200, 170)

		pdf.SetXY(20, 160)
		pdf.SetFont("Helvetica", "", 10)
		pdf.Cell(10, 30, "Notes")

		pdf.SetXY(20, 165)
		pdf.Cell(10, 30, c.Notes)

		/*lastY = 165
		for _, note := range c.Notes {
			pdf.SetXY(20, lastY)
			pdf.Cell(10, 30, note)
			lastY += 5
		}*/
	}

	//
	{
		pdf.SetXY(20, 195)
		pdf.Cell(10, 30, "Banking details")

		pdf.SetXY(20, 200)
		pdf.Cell(10, 30, "VAT")
		pdf.SetXY(30, 200)
		pdf.Cell(10, 30, c.Bank.Vat)

		pdf.SetXY(20, 205)
		pdf.Cell(10, 30, "COC")
		pdf.SetXY(30, 205)
		pdf.Cell(10, 30, c.Bank.Coc)

		pdf.SetXY(20, 210)
		pdf.Cell(10, 30, "IBAN")
		pdf.SetXY(30, 210)
		pdf.Cell(10, 30, c.Bank.Iban)

		/*lastY = 200
		for key, val := range c.Banking {
			pdf.SetXY(20, lastY)
			pdf.Cell(10, 30, key)

			pdf.SetXY(30, lastY)
			pdf.Cell(10, 30, val)
			lastY += 5
		}*/
	}

	pdf.AliasNbPages("{nb}") // replace {nb}
	return pdf, nil
}
