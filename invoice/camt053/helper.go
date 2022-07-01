package camt053

import (
	"io"
)

type PaymentReceived struct {
	Id      string
	Amount  string
	Comment string
	IBAN    string
	Date    string
	Name    string
}

// Helper function to filter out received payments and translate it to
// an easier parsable format.
// WARN: This parser is Rabobank specific
// https://www.rabobank.com/nl/images/formaatbeschrijving-camt-.053-v1.5-rcc.pdf
func FilterPaymentsReceived(r io.Reader) (p []PaymentReceived, e error) {
	e = Read(r, func(head GrpHdr, stmt Stmt) error {
		for _, ntry := range stmt.Ntry {
			if ntry.NtryDtls.TxDtls.BkTxCd.Prtry.Cd == PAYMENT_RECEIVED_RABO {
				p = append(p, PaymentReceived{
					Id:      stmt.Id,
					Amount:  ntry.Amt,
					Comment: ntry.NtryDtls.TxDtls.RmtInf.Ustrd,
					IBAN:    ntry.NtryDtls.TxDtls.RltdPties.DbtrAcct.Id.IBAN,
					Name:    ntry.NtryDtls.TxDtls.RltdPties.Dbtr.Nm,
					Date:    ntry.BookgDt.Dt,
				})
			}
		}
		return nil
	})
	return
}
