package db

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Convert /path/{all}/sub into [/path/a/sub, /path/b/sub]
// Not using https://golang.org/pkg/path/filepath/#Glob so all
// errors are reported
func parseWildcards(paths []string) ([]string, error) {
	var out []string
	for _, path := range paths {
		/* security */
		if !pathFilter(path) {
			return nil, fmt.Errorf("Path hack attempt: %s", path)
		}

		/* all-parser */
		if !strings.Contains(path, "{all}") {
			out = append(out, path)
			continue
		}

		pre := path[:strings.Index(path, "{all}")]
		post := path[len(pre)+len("{all}"):]

		// TODO: future implement?
		if strings.Contains(post, "{all}") {
			return nil, fmt.Errorf("DevErr: Only supporting single {all} in URL")
		}

		founds, e := ioutil.ReadDir(pre)
		if e != nil {
			return nil, e
		}
		for _, found := range founds {
			abs := pre + found.Name() + post
			if f, e := os.Stat(abs); e == nil && f.IsDir() {
				// dir exists
				out = append(out, abs)
			}
		}
	}

	return out, nil
}
