package metrics

import (
	"fmt"
	"net/http"
	"strconv"

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
