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
	Company string
	Entity struct {
		Name string
		Street1 string
		Street2 string
	}
	Customer struct {
		Name string
		Street1 string
		Street2 string
	}
	Meta struct {
		Invoiceid string
		Issuedate string
		Ponumber string
		Duedate string
	}
	Lines []struct {
		Description string
		Quantity string
		Price string
		Total string
	}
	Notes string
	Total struct {
		Ex string
		Tax string
		Total string
	}
	Bank struct {
		Vat string
		Coc string
		Iban string
	}
}