// Package rules contains all domain specific validators
// for this project. (Isolating the regexps here)
package rules

import (
	"gopkg.in/validator.v2"
	"reflect"
	"regexp"
    "fmt"
    "strings"
    "strconv"
)

var regex map[string]*regexp.Regexp

// Pre-compile regexs
func init() {
	regex = make(map[string]*regexp.Regexp)
	regex["slug"] = regexp.MustCompile(`^[a-z0-9_\-]+$`)
    regex["date"] = regexp.MustCompile(`^(19|20)[0-9]{2}-(01|02|03|04|05|06|07|08|09|10|11|12)-(01|02|03|04|05|06|07|08|09|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|25|26|27|28|29|30|31)$`)
    regex["time"] = regexp.MustCompile(`^([0-9]{2}|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`)
    regex["uint"] = regexp.MustCompile(`^[0-9]+$`)

	validator.SetValidationFunc("slug", slug)
    validator.SetValidationFunc("date", date)
    validator.SetValidationFunc("time", time)
    validator.SetValidationFunc("uint", uintFn)
}

func strCheck(rule string, v interface{}, param string) error {
    st := reflect.ValueOf(v)
    if st.Kind() != reflect.String {
        return fmt.Errorf("%s only validates strings", rule)
    }
    if (! regex[rule].MatchString(st.String())) {
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
        var i int
        i, e = strconv.Atoi(strings.Split(st.String(), ":")[0])
        if i > 23 {
            e = fmt.Errorf("date must be in [00-23] range")
        }
    }

    return e
}
func uintFn(v interface{}, param string) error {
    return strCheck("uint", v, param)
}

func Init() error {
	// Do nothing, just get this file included..
    return nil
}