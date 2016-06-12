package invoice

import (
	"database/sql"
)

var db *sql.DB

func Init(d *sql.DB) error {
	db = d
	return nil
}

func getEntity(id int) (*InvoiceEntity, error) {
	c := new(InvoiceEntity)
	e := db.QueryRow(`
		SELECT
			name,
			bank_iban
			bank_bic,
			number_vat,
			number_coc,
			owner_name,
			owner_address_one,
			owner_address_two,
			owner_country
		FROM
			entity
	`).Scan(&c.Name, &c.IBAN, &c.BIC, &c.VAT, &c.COC, &c.Owner, &c.Street1, &c.Street2, &c.Country)
	return c, e
}

func getCustomer(id int) (*InvoiceCustomer, error) {
	c := new(InvoiceCustomer)
	e := db.QueryRow(`
		SELECT
			name,
			address_one,
			address_two,
			country
		FROM
			customer
	`).Scan(&c.Name, &c.Street1, &c.Street2, &c.Country)
	return c, e
}
