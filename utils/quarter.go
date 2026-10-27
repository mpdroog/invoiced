package utils

import (
	"fmt"
	"time"
)

// YearQuarter returns the quarter number (1-4) for a given date.
func YearQuarter(now time.Time) int {
	switch now.Month() {
	case time.January:
		fallthrough
	case time.February:
		fallthrough
	case time.March:
		return 1
	case time.April:
		fallthrough
	case time.May:
		fallthrough
	case time.June:
		return 2
	case time.July:
		fallthrough
	case time.August:
		fallthrough
	case time.September:
		return 3
	case time.October:
		fallthrough
	case time.November:
		fallthrough
	case time.December:
		return 4
	}
	panic(fmt.Sprintf("Invalid month: %d", now.Month()))
}

// QuarterEnd returns the last day of quarter q (1-4) in year.
func QuarterEnd(year, q int) time.Time {
	nextQuarter := time.Date(year, time.Month(3*q+1), 1, 0, 0, 0, 0, time.UTC)
	return nextQuarter.AddDate(0, 0, -1)
}

// CreateInvoiceID creates an invoice ID with YYYY-QN-XXXX pattern.
// For example, 2016Q1-0001 is Invoice 1 in the first Quarter of 2016.
func CreateInvoiceID(now time.Time, idx uint64) string {
	return fmt.Sprintf("%dQ%d-%04d", now.Year(), YearQuarter(now), idx)
}
