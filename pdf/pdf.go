package pdf

import (
	"github.com/jung-kurt/gofpdf"
	"fmt"
)

type Line struct {
	Description string
	Quantity string
	UnitPrice string
	Total string
}

type Content struct {
	CompanyName string
	From []string
	To []string

	Meta map[string]string
	Lines []Line
	TotalEx string
	TotalTax string
	TotalInc string
	Notes []string
	Banking map[string]string
}

func Create(c Content) (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(162, 162, 162);
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
	pdf.AddPage()

	var lastY float64 = 15

	// Company name
	{
		pdf.SetXY(20, 15)
		pdf.SetFont("Helvetica", "B", 18)
		pdf.Cell(10, 30, c.CompanyName)
	}

	// From
	{
		pdf.SetXY(130, 15)
		pdf.SetFont("Helvetica", "", 10)
		pdf.Cell(10, 30, "From")

		lastY = 15
		for _, from := range c.From {
			pdf.SetXY(140, lastY)
			pdf.Cell(10, 30, from)
			lastY += 5
		}
	}

	// Invoice for
	{
		pdf.SetXY(20, 50)
		pdf.SetFont("Helvetica", "", 10)
		pdf.Cell(10, 30, "Invoice For")

		lastY = 50
		for _, to := range c.To {
			pdf.SetXY(40, lastY)
			pdf.Cell(10, 30, to)
			lastY += 5
		}
	}

	// Meta
	{
		lastY = 50
		pdf.SetFont("Helvetica", "", 10)
		for key, val := range c.Meta {
			pdf.SetXY(120, lastY)
			pdf.Cell(10, 30, key)

			pdf.SetXY(143, lastY)
			pdf.Cell(10, 30, val)
			lastY += 5
		}
	}

	// Header
	{
		pdf.SetXY(20, 80)
		pdf.SetFont("Helvetica", "", 10)
		pdf.Cell(10, 30, "Description")

		pdf.SetXY(126, 80)
		pdf.Cell(10, 30, "Quantity")

		pdf.SetXY(150, 80)
		pdf.Cell(10, 30, "Unit Price")

		pdf.SetXY(170, 80)
		pdf.Cell(10, 30, "Amount")

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
			pdf.Cell(10, 30, line.Quantity)

			pdf.SetXY(150, lastY)
			pdf.Cell(10, 30, line.UnitPrice)

			pdf.SetXY(170, lastY)
			pdf.Cell(10, 30, line.Total)

			lastY += 5
		}
	}

	// Totals
	{
		pdf.SetXY(126, 100)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.Cell(10, 30, "Subtotal")

		pdf.SetXY(170, 100)
		pdf.Cell(10, 30, c.TotalEx)

		pdf.SetXY(126, 105)
		pdf.Cell(10, 30, "TAX (21%)")

		pdf.SetXY(170, 105)
		pdf.Cell(10, 30, c.TotalTax)

		pdf.SetXY(126, 120)
		pdf.SetFont("Helvetica", "B", 14)
		pdf.Cell(10, 30, "Amount Due")

		pdf.SetXY(161, 120)
		pdf.Cell(10, 30, string([]byte{byte(128)}) + c.TotalInc)
	}

	// Notes
	{
		pdf.Line(20, 170, 200, 170)

		pdf.SetXY(20, 160)
		pdf.SetFont("Helvetica", "", 10)
		pdf.Cell(10, 30, "Notes")

		lastY = 165
		for _, note := range c.Notes {
			pdf.SetXY(20, lastY)
			pdf.Cell(10, 30, note)
			lastY += 5
		}
	}

	//
	{
		pdf.SetXY(20, 195)
		pdf.Cell(10, 30, "Banking details")

		lastY = 200
		for key, val := range c.Banking {
			pdf.SetXY(20, lastY)
			pdf.Cell(10, 30, key)

			pdf.SetXY(30, lastY)
			pdf.Cell(10, 30, val)
			lastY += 5
		}
	}

	pdf.AliasNbPages("{nb}") // replace {nb}
	return pdf, nil
}