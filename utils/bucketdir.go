// Package utils provides utility functions for path manipulation and ID generation.
package utils

import (
	"strings"
)

// BucketDir extracts the bucket directory name from a path.
func BucketDir(path string) string {
	return strings.SplitN(path, "/", 5)[3]
}
