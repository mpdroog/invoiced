package invoice

/**
  return {
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
    }
  };
*/
type Invoice struct {
	Company string `validate:"nonzero"`
	Entity  struct {
		Name    string `validate:"nonzero"`
		Street1 string `validate:"nonzero"`
		Street2 string `validate:"nonzero"`
	}
	Customer struct {
		Name    string `validate:"nonzero"`
		Street1 string `validate:"nonzero"`
		Street2 string `validate:"nonzero"`
	}
	Meta struct {
		Invoiceid string `validate:"nonzero,slug"`
		Issuedate string `validate:"nonzero,date"`
		Ponumber  string `validate:"slug"`
		Duedate   string `validate:"nonzero,date"`
	}
	Lines []struct {
		Description string `validate:"nonzero"`
		Quantity    string `validate:"nonzero,int"`
		Price       string `validate:"nonzero,price"`
		Total       string `validate:"nonzero,price"`
	}
	Notes string
	Total struct {
		Ex    string `validate:"nonzero,price"`
		Tax   string `validate:"nonzero,price"`
		Total string `validate:"nonzero,price"`
	}
	Bank struct {
		Vat  string `validate:"nonzero"`
		Coc  string `validate:"nonzero"`
		Iban string `validate:"nonzero,iban"`
	}
}
