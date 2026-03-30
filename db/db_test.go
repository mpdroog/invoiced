package db

import (
	"regexp"
	"testing"
)

func TestPathFilter(t *testing.T) {
	// Initialize pathRegex for testing
	pathRegex = regexp.MustCompile(`^[A-Za-z0-9\._\-\/{}]+$`)

	tests := []struct {
		path string
		ok   bool
	}{
		// Valid paths
		{"acme/2024/Q1/file.toml", true},
		{"entity/year/bucket/invoice.toml", true},
		{"test-entity/2024/Q1/sales-invoices-paid/inv.toml", true},
		{"entity_name/file.toml", true},
		{"entity/{all}/invoices", true},

		// Invalid paths - traversal attempts
		{"../etc/passwd", false},
		{"acme/../other/file.toml", false},
		{"acme/2024/../../secret", false},

		// Invalid paths - absolute paths
		{"/etc/passwd", false},
		{"/home/user/file", false},

		// Invalid paths - special characters
		{"acme/file with spaces.toml", false},
		{"acme/file;rm -rf.toml", false},
		{"acme/$HOME/file.toml", false},
		{"acme/`whoami`.toml", false},
	}

	for _, tt := range tests {
		got := pathFilter(tt.path)
		if got != tt.ok {
			t.Errorf("pathFilter(%q) = %v, want %v", tt.path, got, tt.ok)
		}
	}
}
