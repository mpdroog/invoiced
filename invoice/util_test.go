package invoice

import (
	"testing"
	"time"
)

func testYearQuarter(t *testing.T) {
	tests := map[time.Time]int {
		time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC): 1, /* Jan = Q1 */
		time.Date(2016, 2, 1, 0, 0, 0, 0, time.UTC): 1,
		time.Date(2016, 3, 1, 0, 0, 0, 0, time.UTC): 1,
		time.Date(2016, 4, 1, 0, 0, 0, 0, time.UTC): 2, /* April = Q2 */
		time.Date(2016, 5, 1, 0, 0, 0, 0, time.UTC): 2,
		time.Date(2016, 6, 1, 0, 0, 0, 0, time.UTC): 2,
		time.Date(2016, 7, 1, 0, 0, 0, 0, time.UTC): 3, /* July = Q3 */
		time.Date(2016, 8, 1, 0, 0, 0, 0, time.UTC): 3,
		time.Date(2016, 9, 1, 0, 0, 0, 0, time.UTC): 3,
		time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC): 4, /* October = Q4 */
		time.Date(2016, 11, 1, 0, 0, 0, 0, time.UTC): 4,
		time.Date(2016, 12, 1, 0, 0, 0, 0, time.UTC): 4,
	}
	for now, expect := range tests {
		if (yearQuarter(now) != expect) {
			t.Errorf("Date %s expected %d", now.String(), expect)
		}
	}
}

func testCreateInvoiceId(t *testing.T) {
	expect := "2016Q1-0100"
	res := createInvoiceId(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC), 100)
	if expect != res {
		t.Errorf("Invoicepattern mismatch: %s expected, received=%s", expect, res)
	}
}