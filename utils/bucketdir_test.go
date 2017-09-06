package utils

import (
	"testing"
)

// func parentDir(path string) (string, error) {
func TestBucketDir(t *testing.T) {
	fpath := "billingdb/rootdev/2017/Q1/hours/client-2017feb.toml"
	expect := "Q1"

	p := BucketDir(fpath)
	if p != expect {
		t.Errorf("BucketDir(%s) should be %s but is %s", fpath, expect, p)
	}
}