package invoice

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
		Vat     string
		Coc     string
	}
	Meta struct {
		Conceptid string `validate:"slug"`
		Status    string `validate:"slug"`
		Invoiceid string `validate:"slug"`
		Issuedate string `validate:"date"`
		Ponumber  string `validate:"slug"`
		Duedate   string `validate:"nonzero,date"`
		Paydate   string `validate:"date"`
		Freefield string
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
