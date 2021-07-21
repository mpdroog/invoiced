package utils

import (
	"strings"
)

func BucketDir(path string) string {
	return strings.SplitN(path, "/", 5)[3]
}
