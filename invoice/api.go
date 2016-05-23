package invoice

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
)
/*
 config.toml
 - require_login (false)
 - whitelist (127.0.0.1/32)
*/
/*
 config-table
  - invoice_pattern ([YYYYmmdd]-[countyear])
*/
/*
 config
 - key
 - value
 - comment

 tax
 - percent
 - name
 - amount
 - comment

 ledger
 - number
 - name
 - comment

 product
 - date_start
 - date_end
 - date_cancel
 - desciption
 - comment
 - interval
 - ledger
 - tax
 - date_added
 - date_updated

 invoiceline
 - date
 - invoice_id
 - product_id (null)
 - ledger
 - quantity
 - amount
 - sum
 - comment
 - tax
 - date_added
 - date_updated

 invoice
 - id
 - type (INVOICE, CREDITINVOICE, PURCHASE, RECEIPT)
 - files (dir/[year-month]/[customer]/[invoice])
 - related_invoice
 - date_sent
 - date_paid
 - sum
 - txn
 - comment
 - date_added
 - date_updated

 entity
 - type (self, debtor, creditor)
 - parent_entity
 - name
 - zipcode
 - houseno
 - country
 - address
 - comment
 - date_added
 - date_updated

 entity_contact
 - entity_id
 - email
 - type (finance, administration, owner)
 - phone
 - comment
 - date_added
 - date_updated

 login
 - email
 - pass
 - comment
 - date_added
 - date_updated
*/

func List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	//
}