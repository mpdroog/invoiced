package invoice

import (
	"fmt"
	"time"
)

func yearQuarter(now time.Time) int {
	switch now.Month() {
	case time.January:
	case time.February:
	case time.March:
		return 1
	case time.April:
	case time.May:
	case time.June:
		return 2
	case time.July:
	case time.August:
	case time.September:
		return 3
	case time.October:
	case time.November:
	case time.December:
		return 4
	}
	panic(fmt.Sprintf("Invalid month: %d", now.Month()))
}

// Create invoice with YYYY-QN-XXXX pattern
// i.e. 2016-Q1-0001 (Invoice 1 in the first Quarter of 2016)
func createInvoiceId(now time.Time, idx uint64) string {
	return fmt.Sprintf("%dQ%d-%04d", now.Year(), yearQuarter(now), idx)
}