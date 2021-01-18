package config

import (
	"strings"
	"os"
	"fmt"
	"github.com/BurntSushi/toml"
	"log"
)

var (
	Verbose    bool
	DbPath     string
	HTTPListen string
	CurDir     string

	Hostname   string
	C          Config
)

type ConfigQueue struct {
	User     string
	Pass     string
	Host     string
	Port     int
	From     string
	FromReply string
	Display  string
	Subject  string
	BCC      []string
}
type Config struct {
	Queues    map[string]ConfigQueue
}

func Ignore(str string) bool {
	return strings.ToLower(str) == ".ds_store";
}

func Open(f string) error {
	r, e := os.Open(f)
	if e != nil {
		return e
	}
	defer r.Close()
	if _, e := toml.DecodeReader(r, &C); e != nil {
		return fmt.Errorf("TOML: %s", e)
	}

	if Verbose {
		log.Printf("C=%+v\n", C)
	}

	Hostname, e = os.Hostname()
	if e != nil {
		return e
	}
	return nil
}