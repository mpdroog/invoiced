package taxes

// Dutch tax rules (Belastingdienst)

import (
	"regexp"
	"time"

	"github.com/mpdroog/invoiced/utils"
)

var (
	// VAT-return (btw-aangifte) PDF as downloaded from the Belastingdienst, incl. older naming
	filedRegex = regexp.MustCompile(`(?i)^(btw-aangifte.*|aangifte(_omzetbelasting.*)?|ob|belastingdienst-ob.*)\.pdf$`)
	// ICP-return (opgaaf intracommunautaire prestaties)
	icpRegex = regexp.MustCompile(`(?i)^(opgaaf_icp.*|icp|belastingdienst-icp.*)\.pdf$`)
	// Payment receipt (iDEAL)
	paidRegex = regexp.MustCompile(`(?i)^online_betalen_aan_de_belastingdienst.*\.pdf$`)
)

// BTWDeadline returns the last day to file and pay the btw-aangifte of quarter q:
// the last day of the month after the quarter.
func BTWDeadline(year, q int) time.Time {
	nextMonth := utils.QuarterEnd(year, q).AddDate(0, 0, 1)
	return nextMonth.AddDate(0, 1, -1)
}
