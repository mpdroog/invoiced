// Package rules contains all domain specific validators
// for this project. (Isolating the regexps here)
package rules

import (
	"fmt"
	"gopkg.in/validator.v2"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var regex map[string]*regexp.Regexp

// Pre-compile regexs
func init() {
	regex = make(map[string]*regexp.Regexp)
	regex["slug"] = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
	regex["date"] = regexp.MustCompile(`^(19|20)[0-9]{2}-(01|02|03|04|05|06|07|08|09|10|11|12)-(01|02|03|04|05|06|07|08|09|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|25|26|27|28|29|30|31)$`)
	regex["time"] = regexp.MustCompile(`^([0-9]{2}|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`)
	regex["uint"] = regexp.MustCompile(`^[0-9]+$`)
	regex["qty"] = regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
	regex["price"] = regexp.MustCompile(`^-?[0-9]+\.[0-9]+$`)

	validator.SetValidationFunc("slug", slug)
	validator.SetValidationFunc("date", date)
	validator.SetValidationFunc("time", time)
	validator.SetValidationFunc("uint", uintFn)
	validator.SetValidationFunc("iban", iban)
	validator.SetValidationFunc("price", price)
	validator.SetValidationFunc("qty", qty)
}

func strCheck(rule string, v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return fmt.Errorf("%s only validates strings", rule)
	}
	if st.String() == "" {
		// Support optional fields
		return nil
	}
	if !regex[rule].MatchString(st.String()) {
		return fmt.Errorf("%s did not match string", rule)
	}
	return nil
}

func slug(v interface{}, param string) error {
	return strCheck("slug", v, param)
}
func date(v interface{}, param string) error {
	return strCheck("date", v, param)
}
func time(v interface{}, param string) error {
	e := strCheck("time", v, param)

	if e == nil {
		// Post-validate for > 23:59
		st := reflect.ValueOf(v)
		if st.String() != "" {
			var i int
			i, e = strconv.Atoi(strings.Split(st.String(), ":")[0])
			if i > 23 {
				e = fmt.Errorf("date must be in [00-23] range")
			}
		}
	}

	return e
}
func uintFn(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() == reflect.String {
		return strCheck("uint", v, param)
	}
	if st.Kind() == reflect.Uint {
		return nil
	}
	if st.Kind() == reflect.Float64 {
		n := st.Float();
		if n <= 0 {
			return fmt.Errorf("uint negative not allowed. Given="+st.String())
		}
		return nil
	}
	return fmt.Errorf("uint no validator for reflect kind=%s", st.Kind())
}
func price(v interface{}, param string) error {
	return strCheck("price", v, param)
}
func qty(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() == reflect.String {
		return strCheck("qty", v, param)
	}
	if st.Kind() == reflect.Uint {
		return nil
	}
	if st.Kind() == reflect.Float64 {
		n := st.Float();
		if n <= 0 {
			return fmt.Errorf("qty negative not allowed. Given="+st.String())
		}
		return nil
	}
	return fmt.Errorf("qty no validator for reflect kind=%s", st.Kind())
}
func Init() error {
	// Do nothing, just get this file included..
	return nil
}
