package camt053

// PaymentReceivedRabo is the Rabobank transaction code for received payments.
const PaymentReceivedRabo = "541"

// GrpHdr represents the CAMT053 group header element.
type GrpHdr struct {
	MsgID   string `xml:"MsgId"`
	CreDtTm string `xml:"CreDtTm"`
}

// Stmt represents a CAMT053 statement element.
type Stmt struct {
	ID           string `xml:"Id"`
	ElctrncSeqNb int    `xml:"ElctrncSeqNb"`
	CreDtTm      string `xml:"CreDtTm"`
	Acct         struct {
		ID struct {
			IBAN string `xml:"IBAN"`
		} `xml:"Id"`
		Ccy string `xml:"Ccy"`
	} `xml:"Acct"`
	Bal []struct {
		Tp struct {
			CdOrPrtry struct {
				Cd string `xml:"Cd"`
			} `xml:"CdOrPrtry"`
		} `xml:"Tp"`
		Amt       string `xml:"Amt"`
		CdtDbtInd string `xml:"CdtDbtInd"`
		Dt        struct {
			Dt string `xml:"Dt"`
		} `xml:"Dt"`
	} `xml:"Bal"`
	TxsSummry struct {
		TtlNtries struct { //nolint:revive // field names match XML schema
			NbOfNtries    int     `xml:"NbOfNtries"`
			Sum           float64 `xml:"Sum"`
			TtlNetNtryAmt float64 `xml:"TtlNetNtryAmt"` //nolint:revive // field name matches XML schema
			CdtDbtInd     string  `xml:"CdtDbtInd"`
		} `xml:"TtlNtries"`
	} `xml:"TxsSummry"`

	Ntry []struct {
		Amt       string `xml:"Amt"`
		CdtDbtInd string `xml:"CdtDbtInd"`
		RvslInd   string `xml:"RvslInd"`
		Sts       string `xml:"Sts"`
		BookgDt   struct {
			Dt string `xml:"Dt"`
		} `xml:"BookgDt"`
		ValDt struct {
			Dt string `xml:"Dt"`
		} `xml:"ValDt"`
		BkTxCd struct {
			Prtry struct {
				Cd string `xml:"Cd"`
			} `xml:"Prtry"`
		} `xml:"BkTxCd"`
		NtryDtls struct {
			TxDtls struct {
				BkTxCd struct {
					Prtry struct {
						Cd string `xml:"Cd"`
					} `xml:"Prtry"`
				} `xml:"BkTxCd"`
				RltdPties struct {
					Cdtr struct {
						Nm      string `xml:"Nm"`
						PstlAdr struct {
							AdrTp   string   `xml:"AdrTp"`
							Ctry    string   `xml:"Ctry"`
							AdrLine []string `xml:"AdrLine"`
						} `xml:"PstlAdr"`
					} `xml:"Cdtr"`
					CdtrAcct struct {
						ID struct {
							IBAN string `xml:"IBAN"`
						} `xml:"Id"`
					} `xml:"CdtrAcct"`

					Dbtr struct {
						Nm      string `xml:"Nm"`
						PstlAdr struct {
							AdrTp   string   `xml:"AdrTp"`
							Ctry    string   `xml:"Ctry"`
							AdrLine []string `xml:"AdrLine"`
						} `xml:"PstlAdr"`
					} `xml:"Dbtr"`
					DbtrAcct struct {
						ID struct {
							IBAN string `xml:"IBAN"`
						} `xml:"Id"`
					} `xml:"DbtrAcct"`
				} `xml:"RltdPties"`
				RltdAgts struct {
					CdtrAgt struct {
						FinInstnID struct {
							BIC string `xml:"BIC"`
						} `xml:"FinInstnId"`
					} `xml:"CdtrAgt"`
				} `xml:"RltdAgts"`
				RmtInf struct {
					Ustrd string `xml:"Ustrd"`
					Strd  struct {
						CdtrRefInf struct {
							Tp struct {
								CdOrPrtry struct {
									Cd string `xml:"Cd"`
								} `xml:"CdOrPrtry"`
								Issr string `xml:"Issr"`
							} `xml:"Tp"`
							Ref string `xml:"Ref"`
						} `xml:"CdtrRefInf"`
					} `xml:"Strd"`
				} `xml:"RmtInf"`
			} `xml:"TxDtls"`
		} `xml:"NtryDtls"`
	} `xml:"Ntry"`
}
