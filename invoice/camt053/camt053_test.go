package camt053

import (
	"bufio"
	"os"
	"testing"
)

func TestRead(t *testing.T) {
	f, e := os.Open("./CAMT053.xml")
	if e != nil {
		t.Fatal(e)
	}
	buf := bufio.NewReader(f)

	if e := Read(buf, func(head GrpHdr, stmt Stmt) error {
		/*for _, ntry := range stmt.Ntry {
			if ntry.NtryDtls.TxDtls.BkTxCd.Prtry.Cd == PAYMENT_RECEIVED_RABO {
				fmt.Printf(
					"%s %s %s %sEUR %s %s %s - %s (%s)\n",
					stmt.Id, ntry.NtryDtls.TxDtls.RltdPties.Cdtr.Nm, ntry.NtryDtls.TxDtls.RltdPties.CdtrAcct.Id.IBAN,
					ntry.Amt, ntry.CdtDbtInd,
					ntry.NtryDtls.TxDtls.RmtInf.Ustrd,
					ntry.BookgDt.Dt,

					ntry.NtryDtls.TxDtls.RltdPties.Dbtr.Nm,
					ntry.NtryDtls.TxDtls.BkTxCd.Prtry.Cd,
				)
			}
		}*/

		if stmt.ElctrncSeqNb < 16141 || stmt.ElctrncSeqNb > 16206 {
			t.Fatalf("Invalid ElctrncSeqNb")
		}
		if stmt.Acct.Id.IBAN != "NL59RABO3181240869" && stmt.Acct.Id.IBAN != "NL17RABO0310029597" {
			t.Fatalf("Invalid IBAN=" + stmt.Acct.Id.IBAN)
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
}

func TestFilterPaymentsReceived(t *testing.T) {
	f, e := os.Open("./CAMT053.xml")
	if e != nil {
		t.Fatal(e)
	}
	buf := bufio.NewReader(f)
	p, e := FilterPaymentsReceived(buf)
	if e != nil {
		t.Fatal(e)
	}

	if len(p) != 3 {
		t.Errorf("Payments received != 3", len(p))
	}
	if p[0].IBAN != "NL62ABNA0408441224" {
		t.Errorf("Payment(%s) invalid IBAN=%s", p[0].Id, p[0].IBAN)
	}
	if p[0].Amount != "4425.25" {
		t.Errorf("Payment(%s) invalid amount=%s", p[0].Id, p[0].Amount)
	}
	if p[0].Comment != "2016Q3-0004" {
		t.Errorf("Payment(%s) invalid comment=%s", p[0].Id, p[0].Comment)
	}
	if p[0].Date != "2016-08-15" {
		t.Errorf("Payment(%s) invalid date=%s", p[0].Id, p[0].Date)
	}
	if p[0].Name != "XS NEWS B V" {
		t.Errorf("Payment(%s) invalid name=%s", p[0].Id, p[0].Name)
	}
}
