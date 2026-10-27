package utils

import (
	"testing"
	"time"
)

func TestYearQuarter(t *testing.T) {
	tests := map[time.Time]int{
		time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC):  1, /* Jan = Q1 */
		time.Date(2016, 2, 1, 0, 0, 0, 0, time.UTC):  1,
		time.Date(2016, 3, 1, 0, 0, 0, 0, time.UTC):  1,
		time.Date(2016, 4, 1, 0, 0, 0, 0, time.UTC):  2, /* April = Q2 */
		time.Date(2016, 5, 1, 0, 0, 0, 0, time.UTC):  2,
		time.Date(2016, 6, 1, 0, 0, 0, 0, time.UTC):  2,
		time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC):  3, /* July = Q3 */
		time.Date(2016, 8, 1, 0, 0, 0, 0, time.UTC):  3,
		time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC):  3,
		time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC): 4, /* October = Q4 */
		time.Date(2016, 11, 1, 0, 0, 0, 0, time.UTC): 4,
		time.Date(2016, 12, 1, 0, 0, 0, 0, time.UTC): 4,
	}
	for now, expect := range tests {
		if YearQuarter(now) != expect {
			t.Errorf("Date %s expected %d", now.String(), expect)
		}
	}
}

func TestCreateInvoiceID(t *testing.T) {
	expect := "2016Q1-0100"
	res := CreateInvoiceID(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC), 100)
	if expect != res {
		t.Errorf("Invoicepattern mismatch: %s expected, received=%s", expect, res)
	}
}

func TestQuarterEnd(t *testing.T) {
	tests := map[int]string{
		1: "2024-03-31",
		2: "2024-06-30",
		3: "2024-09-30",
		4: "2024-12-31",
	}
	for q, expect := range tests {
		if res := QuarterEnd(2024, q).Format("2006-01-02"); res != expect {
			t.Errorf("Q%d expected %s, received=%s", q, expect, res)
		}
	}
}
