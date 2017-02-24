package rules

import (
	"testing"
)

func TestIban(t *testing.T) {
	valid := []string{
		"AL47212110090000000235698741",
		"AZ21NABZ00000000137010001944",
		"AT611904300234573201",
		"BE68539007547034",
		"DE89370400440532013000",
		"FR1420041010050500013M02606",
		"NL91ABNA0417164300",
	}
	invalid := []string{
		"HE LLO",
		"HELLO",
		"NL1234",
		"-999",
		"-0",
		"NL91ABNA041716430",
		"GB00",
	}

	for _, str := range valid {
		if e := iban(str, ""); e != nil {
			t.Errorf("match %s failed with e=%s", str, e)
		}
	}
	for _, str := range invalid {
		if e := iban(str, ""); e == nil {
			t.Errorf("str should fail but didn't, input=%s", str)
		}
	}
}
