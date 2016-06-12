package invoice

/*
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
  issuedate: "",
  ponumber: "",
  duedate: ""
},
lines: [{
  description: "",
  quantity: 1,
  price: "12.00",
  total: "12.00"
}],
notes: "",
total: {
  ex: "",
  tax: "",
  total: ""
},
bank: {
  vat: "",
  coc: "",
  iban: ""
}
*/
type InvoiceEntity struct {
	Name string
	Owner string
	Street1 string
	Street2 string
	VAT string
	COC string
	IBAN string
	BIC string
	Country string
}
type InvoiceCustomer struct {
	Name string
	Street1 string
	Street2 string
	Country string
}
type InvoiceMeta struct {
	Invoiceid string
	Issuedate string
	Ponumber string
	Duedate string

	Notes string
	TotalEx float64
	TotalTax float64
	TotalSum float64
}
type InvoiceLine struct {
	Description string
	Quantity int
	Price float64
	Total float64
}

type Invoice struct {
	Company string
	Entity InvoiceEntity
	Customer InvoiceCustomer
	Meta InvoiceMeta
	Lines []InvoiceLine
}