package camt053

import (
	"io"
)

// PaymentReceived represents a received payment parsed from a CAMT053 statement.
type PaymentReceived struct {
	ID      string
	Amount  string
	Comment string
	IBAN    string
	Date    string
	Name    string
}

// FilterPaymentsReceived filters received payments from a CAMT053 file and
// translates them to an easier parsable format.
// WARN: This parser is Rabobank specific
// https://www.rabobank.com/nl/images/formaatbeschrijving-camt-.053-v1.5-rcc.pdf
func FilterPaymentsReceived(r io.Reader) (p []PaymentReceived, e error) {
	e = Read(r, func(_ GrpHdr, stmt Stmt) error {
		for _, ntry := range stmt.Ntry {
			if ntry.NtryDtls.TxDtls.BkTxCd.Prtry.Cd == PaymentReceivedRabo {
				p = append(p, PaymentReceived{
					ID:      stmt.ID,
					Amount:  ntry.Amt,
					Comment: ntry.NtryDtls.TxDtls.RmtInf.Ustrd,
					IBAN:    ntry.NtryDtls.TxDtls.RltdPties.DbtrAcct.ID.IBAN,
					Name:    ntry.NtryDtls.TxDtls.RltdPties.Dbtr.Nm,
					Date:    ntry.BookgDt.Dt,
				})
			}
		}
		return nil
	})
	return
}
