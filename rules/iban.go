package rules

import (
	"fmt"
	check "github.com/go-pascal/iban"
	"reflect"
)

func iban(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return fmt.Errorf("iban only validates strings")
	}
	if st.String() == "" {
		// Support optional fields
		return nil
	}

	ok, _, e := check.IsCorrectIban(st.String(), true)
	if e != nil {
		return e
	}
	if !ok {
		return fmt.Errorf("iban not valid")
	}
	return nil
}
