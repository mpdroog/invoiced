package taxes

import (
	"testing"
)

func TestBTWDeadline(t *testing.T) {
	tests := map[int]string{
		1: "2024-04-30",
		2: "2024-07-31",
		3: "2024-10-31",
		4: "2025-01-31", /* Q4 deadline is next year */
	}
	for q, expect := range tests {
		if res := BTWDeadline(2024, q).Format("2006-01-02"); res != expect {
			t.Errorf("Q%d expected %s, received=%s", q, expect, res)
		}
	}
}

func TestFiledRegex(t *testing.T) {
	tests := map[string]bool{
		"Btw-aangifte__2025_09-04-2025_18.42u.pdf":           true,
		"Aangifte_omzetbelasting_2021_03-01-2022_16.08u.pdf": true,
		"Aangifte.pdf":                         true,
		"OB.pdf":                               true,
		"Belastingdienst-OB-2018Q1.pdf":        true,
		"Aangifte_inkomstenbelasting_2024.pdf": false,
		"aangifte-icp.pdf":                     false,
		"invoice_25762854.pdf":                 false,
	}
	for name, expect := range tests {
		if res := filedRegex.MatchString(name); res != expect {
			t.Errorf("%s expected %v, received=%v", name, expect, res)
		}
	}
}
