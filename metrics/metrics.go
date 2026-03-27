package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mpdroog/invoiced/config"
	"github.com/mpdroog/invoiced/idx"
	"github.com/mpdroog/invoiced/writer"
	"log"
)

type DashboardMetric struct {
	RevenueTotal string
	RevenueEx    string
	Hours        string
}

// DashboardResponse contains all dashboard data
type DashboardResponse struct {
	// Monthly metrics (existing)
	Monthly map[string]*idx.MonthlyMetric `json:"monthly"`

	// Previous year monthly metrics for comparison
	MonthlyPrevYear map[string]*idx.MonthlyMetric `json:"monthlyPrevYear"`

	// Unpaid invoices summary
	Unpaid *idx.UnpaidSummary `json:"unpaid"`

	// Overdue invoices
	Overdue []idx.OverdueInvoice `json:"overdue"`

	// Quarterly breakdown
	Quarters []idx.QuarterSummary `json:"quarters"`

	// Unbilled hours
	UnbilledHours *idx.UnbilledHoursSummary `json:"unbilledHours"`

	// Year comparison
	YearComparison *idx.YearComparison `json:"yearComparison"`

	// Top clients
	TopClients []idx.CustomerTotal `json:"topClients"`
}

func Dashboard(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year, err := strconv.Atoi(ps.ByName("year"))
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.Dashboard: invalid year: %s", err.Error()), 400)
		return
	}

	if idx.DB == nil {
		http.Error(w, "Index not initialized", 500)
		return
	}

	idxMetrics, err := idx.GetMonthlyMetrics(entity, year)
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.Dashboard: %s", err.Error()), 500)
		return
	}

	// Convert to response format
	m := make(map[string]*DashboardMetric)
	for yearmonth, metric := range idxMetrics {
		m[yearmonth] = &DashboardMetric{
			RevenueTotal: metric.RevenueTotal,
			RevenueEx:    metric.RevenueEx,
			Hours:        metric.Hours,
		}
	}

	if config.Verbose {
		log.Printf("metrics.Dashboard count=%d", len(m))
	}

	if err := writer.Encode(w, r, m); err != nil {
		http.Error(w, fmt.Sprintf("metrics.Dashboard: encode failed: %s", err.Error()), 500)
		return
	}
}

// DashboardFull returns comprehensive dashboard data
func DashboardFull(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	entity := ps.ByName("entity")
	year, err := strconv.Atoi(ps.ByName("year"))
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.DashboardFull: invalid year: %s", err.Error()), 400)
		return
	}

	if idx.DB == nil {
		http.Error(w, "Index not initialized", 500)
		return
	}

	today := time.Now().Format("2006-01-02")

	resp := &DashboardResponse{}

	// Monthly metrics
	monthly, err := idx.GetMonthlyMetrics(entity, year)
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.DashboardFull: monthly: %s", err.Error()), 500)
		return
	}
	resp.Monthly = monthly

	// Previous year monthly metrics for comparison
	monthlyPrev, err := idx.GetMonthlyMetrics(entity, year-1)
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.DashboardFull: monthlyPrev: %s", err.Error()), 500)
		return
	}
	resp.MonthlyPrevYear = monthlyPrev

	// Unpaid summary
	unpaid, err := idx.GetUnpaidSummary(entity, year)
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.DashboardFull: unpaid: %s", err.Error()), 500)
		return
	}
	resp.Unpaid = unpaid

	// Overdue invoices
	overdue, err := idx.GetOverdueInvoices(entity, year, today)
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.DashboardFull: overdue: %s", err.Error()), 500)
		return
	}
	resp.Overdue = overdue

	// Quarterly breakdown
	quarters, err := idx.GetYearlyQuarterSummary(entity, year)
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.DashboardFull: quarters: %s", err.Error()), 500)
		return
	}
	resp.Quarters = quarters

	// Unbilled hours
	unbilled, err := idx.GetUnbilledHours(entity, year)
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.DashboardFull: unbilled: %s", err.Error()), 500)
		return
	}
	resp.UnbilledHours = unbilled

	// Year comparison
	comparison, err := idx.GetYearComparison(entity, year)
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.DashboardFull: comparison: %s", err.Error()), 500)
		return
	}
	resp.YearComparison = comparison

	// Top clients (limit to 5)
	clients, err := idx.GetYearlyCustomerTotals(entity, year, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("metrics.DashboardFull: clients: %s", err.Error()), 500)
		return
	}
	if len(clients) > 5 {
		clients = clients[:5]
	}
	resp.TopClients = clients

	if config.Verbose {
		log.Printf("metrics.DashboardFull entity=%s year=%d", entity, year)
	}

	if err := writer.Encode(w, r, resp); err != nil {
		http.Error(w, fmt.Sprintf("metrics.DashboardFull: encode failed: %s", err.Error()), 500)
		return
	}
}
