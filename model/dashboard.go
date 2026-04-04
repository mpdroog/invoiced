// Package model defines shared data structures for invoices, hours, and taxes.
package model

// TaxSummary contains aggregated tax data for a quarter
type TaxSummary struct {
	Ex        string            // Sum revenue of NL invoices
	Tax       string            // Tax to pay
	EUEx      string            // Sum revenue of EU invoices
	EUCompany map[string]string // Revenue per EU company VAT number (for ICP)
	ExWorld   string            // Sum revenue of world invoices
	ExRevenue string            // Sum revenue of everything
}

// InvoiceAuditLine contains info for audit logging
type InvoiceAuditLine struct {
	InvoiceID   string
	TaxCategory string
	TotalEx     string
	TotalTax    string
}

// CustomerTotal represents revenue per customer
type CustomerTotal struct {
	Name         string
	Revenue      string
	InvoiceCount int
}

// QuarterSummary contains aggregated data for a quarter
type QuarterSummary struct {
	Quarter       int
	InvoiceCount  int
	TotalRevenue  string
	TotalTax      string
	PaidCount     int
	UnpaidCount   int
	PaidRevenue   string
	UnpaidRevenue string
}

// MonthlyMetric contains revenue and hours for a month
type MonthlyMetric struct {
	RevenueTotal string
	RevenueEx    string
	Hours        string
}

// UnpaidSummary contains summary of unpaid invoices
type UnpaidSummary struct {
	Count       int
	TotalAmount string
}

// OverdueInvoice represents an overdue invoice
type OverdueInvoice struct {
	ID           string
	InvoiceID    string
	CustomerName string
	DueDate      string
	Amount       string
	DaysOverdue  int
	Quarter      int
}

// UnbilledHoursSummary contains summary of unbilled hours
type UnbilledHoursSummary struct {
	Count      int
	TotalHours string
}

// YearComparison contains comparison between two years
type YearComparison struct {
	CurrentYear     int
	PreviousYear    int
	CurrentRevenue  string
	PreviousRevenue string
	GrowthPercent   string
	GrowthAmount    string
}

// DashboardMetric is used by the simple dashboard endpoint
type DashboardMetric struct {
	RevenueTotal string
	RevenueEx    string
	Hours        string
}

// DashboardResponse contains all dashboard data for the API
type DashboardResponse struct {
	Monthly         map[string]*MonthlyMetric `json:"monthly"`
	MonthlyPrevYear map[string]*MonthlyMetric `json:"monthlyPrevYear"`
	Unpaid          UnpaidSummary             `json:"unpaid"`
	Overdue         []OverdueInvoice          `json:"overdue"`
	Quarters        []QuarterSummary          `json:"quarters"`
	UnbilledHours   UnbilledHoursSummary      `json:"unbilledHours"`
	YearComparison  YearComparison            `json:"yearComparison"`
	TopClients      []CustomerTotal           `json:"topClients"`
}

// AccountingInvoice contains invoice data for accounting export
type AccountingInvoice struct {
	InvoiceID      string
	Issuedate      string
	CustomerName   string
	CustomerVAT    string
	TaxCategory    string // NL, EU0, WORLD0
	Status         string // PAID, UNPAID
	Quarter        int
	TotalEx        float64
	TotalTax       float64
	TotalInc       float64
	AccountingCode string // From debtor record
}

// AccountingCompany contains company totals for accounting export
type AccountingCompany struct {
	Name           string
	VAT            string
	TaxCategory    string // NL, EU0, WORLD0
	TotalRevenue   float64
	AccountingCode string
}

// AccountingExport contains all data for the accounting Excel export
type AccountingExport struct {
	Invoices     []AccountingInvoice
	Companies    []AccountingCompany
	TotalRevenue float64
	TotalEx      float64
	TotalTax     float64
	TotalHours   float64
}
