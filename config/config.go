package config

import (
	"strings"
)

var (
	Verbose    bool
	DbPath     string
	HTTPListen string
	CurDir     string
	HTTPSOnly  bool
	Local      bool
)

func Ignore(str string) bool {
	return strings.ToLower(str) == ".ds_store";
}