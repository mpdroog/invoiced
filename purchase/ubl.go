package purchase

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
)

// UBLInvoice represents a UBL 2.1 invoice XML structure.
type UBLInvoice struct {
	XMLName      xml.Name         `xml:"Invoice"`
	ID           string           `xml:"ID"`
	IssueDate    string           `xml:"IssueDate"`
	DueDate      string           `xml:"DueDate"`
	Currency     string           `xml:"DocumentCurrencyCode"`
	BuyerRef     string           `xml:"BuyerReference"`
	Supplier     UBLParty         `xml:"AccountingSupplierParty>Party"`
	Customer     UBLParty         `xml:"AccountingCustomerParty>Party"`
	PaymentMeans UBLPaymentMeans  `xml:"PaymentMeans"`
	TaxTotal     UBLTaxTotal      `xml:"TaxTotal"`
	LegalTotal   UBLLegalTotal    `xml:"LegalMonetaryTotal"`
	Lines        []UBLInvoiceLine `xml:"InvoiceLine"`
	Attachment   UBLAttachment    `xml:"AdditionalDocumentReference"`
}

// UBLParty represents party information in a UBL invoice.
type UBLParty struct {
	Name  string `xml:"PartyName>Name"`
	VAT   string `xml:"PartyTaxScheme>CompanyID"`
	COC   string `xml:"PartyLegalEntity>CompanyID"`
	Email string `xml:"Contact>ElectronicMail"`
}

// UBLPaymentMeans represents payment information in a UBL invoice.
type UBLPaymentMeans struct {
	DueDate   string `xml:"PaymentDueDate"`
	PaymentID string `xml:"PaymentID"`
	IBAN      string `xml:"PayeeFinancialAccount>ID"`
	BIC       string `xml:"PayeeFinancialAccount>FinancialInstitutionBranch>FinancialInstitution>ID"`
}

// UBLTaxTotal represents tax totals in a UBL invoice.
type UBLTaxTotal struct {
	TaxAmount string `xml:"TaxAmount"`
}

// UBLLegalTotal represents monetary totals in a UBL invoice.
type UBLLegalTotal struct {
	LineExtension string `xml:"LineExtensionAmount"`
	TaxExclusive  string `xml:"TaxExclusiveAmount"`
	TaxInclusive  string `xml:"TaxInclusiveAmount"`
	Payable       string `xml:"PayableAmount"`
}

// UBLInvoiceLine represents a line item in a UBL invoice.
type UBLInvoiceLine struct {
	ID          string `xml:"ID"`
	Quantity    string `xml:"InvoicedQuantity"`
	LineTotal   string `xml:"LineExtensionAmount"`
	Description string `xml:"Item>Description"`
	Name        string `xml:"Item>Name"`
	TaxPercent  string `xml:"Item>ClassifiedTaxCategory>Percent"`
	Price       string `xml:"Price>PriceAmount"`
}

// UBLAttachment represents an attachment reference in a UBL invoice.
type UBLAttachment struct {
	ID           string              `xml:"ID"`
	DocumentType string              `xml:"DocumentType"`
	Attachment   UBLEmbeddedDocument `xml:"Attachment"`
}

// UBLEmbeddedDocument represents an embedded document in a UBL invoice.
type UBLEmbeddedDocument struct {
	BinaryObject UBLBinaryObject `xml:"EmbeddedDocumentBinaryObject"`
}

// UBLBinaryObject represents a binary object in a UBL invoice attachment.
type UBLBinaryObject struct {
	MimeCode string `xml:"mimeCode,attr"`
	Filename string `xml:"filename,attr"`
	Data     string `xml:",chardata"`
}

// ParseUBL parses a UBL invoice XML and returns a PurchaseInvoice
func ParseUBL(r io.Reader) (*PurchaseInvoice, []byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("read XML: %w", err)
	}

	var ubl UBLInvoice
	if err := xml.Unmarshal(data, &ubl); err != nil {
		return nil, nil, fmt.Errorf("parse XML: %w", err)
	}

	inv := &PurchaseInvoice{
		ID:         ubl.ID,
		Issuedate:  ubl.IssueDate,
		Duedate:    ubl.DueDate,
		Currency:   ubl.Currency,
		PaymentRef: ubl.PaymentMeans.PaymentID,
		IBAN:       ubl.PaymentMeans.IBAN,
		BIC:        ubl.PaymentMeans.BIC,
		TotalEx:    ubl.LegalTotal.TaxExclusive,
		TotalTax:   ubl.TaxTotal.TaxAmount,
		TotalInc:   ubl.LegalTotal.TaxInclusive,
		Status:     "UNPAID",
		Supplier: Supplier{
			Name:  ubl.Supplier.Name,
			VAT:   ubl.Supplier.VAT,
			COC:   ubl.Supplier.COC,
			Email: ubl.Supplier.Email,
		},
	}

	// Parse invoice lines
	for _, line := range ubl.Lines {
		inv.Lines = append(inv.Lines, PurchaseLine{
			Description: line.Description,
			Quantity:    line.Quantity,
			Price:       line.Price,
			Total:       line.LineTotal,
			TaxPercent:  line.TaxPercent,
		})
	}

	// Extract embedded PDF if present
	var pdfData []byte
	binaryData := ubl.Attachment.Attachment.BinaryObject.Data
	if binaryData != "" && ubl.Attachment.DocumentType == "PrimaryImage" {
		inv.PDFFilename = ubl.Attachment.Attachment.BinaryObject.Filename
		if inv.PDFFilename == "" {
			inv.PDFFilename = ubl.Attachment.ID
		}
		var err error
		pdfData, err = base64.StdEncoding.DecodeString(binaryData)
		if err != nil {
			return nil, nil, fmt.Errorf("decode PDF: %w", err)
		}
	}

	return inv, pdfData, nil
}
