package rules

import (
	"testing"
)

func TestSlug(t *testing.T) {
	valid := []string {
		"hello",
		"hel-lo",
		"hel_lo",
		"",
	}
	invalid := []string {
		"HE LLO",
		",HEL,LO",
		"±HI",
	}

	for _, str := range valid {
		if e := slug(str, ""); e != nil {
			t.Errorf("match %s failed with e=%s", str, e)
		}
	}
	for _, str := range invalid {
		if e := slug(str, ""); e == nil {
			t.Errorf("str should fail but didn't, input="+str)
		}
	}
}

func TestDate(t *testing.T) {
	valid := []string {
		"2016-01-01",
		"2016-12-01",
		"1980-01-01",
		"2099-12-31",
		"",
	}
	invalid := []string {
		"HE LLO",
		"HELLO",
		"2016-02-HELLO",
		"2016-12-1",
		"2016-1-12",
		"2016-13-01",
		"2016-10-33",
		"1830-10-33",
		"3016-10-33",
	}

	for _, str := range valid {
		if e := date(str, ""); e != nil {
			t.Errorf("match %s failed with e=%s", str, e)
		}
	}
	for _, str := range invalid {
		if e := date(str, ""); e == nil {
			t.Errorf("str should fail but didn't, input=%s", str)
		}
	}
}

func TestTime(t *testing.T) {
	valid := []string {
		"00:00",
		"01:00",
		"12:59",
		"23:59",
		"",
	}
	invalid := []string {
		"HE LLO",
		"HELLO",
		"1:01",
		"24:00",
		"99:00",
		"00:60",
	}

	for _, str := range valid {
		if e := time(str, ""); e != nil {
			t.Errorf("match %s failed with e=%s", str, e)
		}
	}
	for _, str := range invalid {
		if e := time(str, ""); e == nil {
			t.Errorf("str should fail but didn't, input=%s", str)
		}
	}
}

func TestUint(t *testing.T) {
	valid := []string {
		"12",
		"0",
		"99",
		"9999999999999",
		"",
	}
	invalid := []string {
		"HE LLO",
		"HELLO",
		"-10",
		"-999",
		"-0",
	}

	for _, str := range valid {
		if e := uintFn(str, ""); e != nil {
			t.Errorf("match %s failed with e=%s", str, e)
		}
	}
	for _, str := range invalid {
		if e := uintFn(str, ""); e == nil {
			t.Errorf("str should fail but didn't, input=%s", str)
		}
	}
}
