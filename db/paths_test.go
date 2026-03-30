package db

import (
	"testing"
)

func TestInvoicePath(t *testing.T) {
	tests := []struct {
		entity, year, bucket, name string
		paid                       bool
		want                       string
	}{
		{"acme", "2024", "Q1", "2024Q1-0001", false, "acme/2024/Q1/sales-invoices-unpaid/2024Q1-0001.toml"},
		{"acme", "2024", "Q1", "2024Q1-0001", true, "acme/2024/Q1/sales-invoices-paid/2024Q1-0001.toml"},
		{"corp", "2023", "Q4", "2023Q4-0099", false, "corp/2023/Q4/sales-invoices-unpaid/2023Q4-0099.toml"},
	}

	for _, tt := range tests {
		got := InvoicePath(tt.entity, tt.year, tt.bucket, tt.name, tt.paid)
		if got != tt.want {
			t.Errorf("InvoicePath(%q, %q, %q, %q, %v) = %q, want %q",
				tt.entity, tt.year, tt.bucket, tt.name, tt.paid, got, tt.want)
		}
	}
}

func TestConceptInvoicePath(t *testing.T) {
	got := ConceptInvoicePath("acme", "2024", "abc123")
	want := "acme/2024/concepts/sales-invoices/abc123.toml"
	if got != want {
		t.Errorf("ConceptInvoicePath() = %q, want %q", got, want)
	}
}

func TestHourPath(t *testing.T) {
	got := HourPath("acme", "2024", "Q2", "project-hours")
	want := "acme/2024/Q2/hours/project-hours.toml"
	if got != want {
		t.Errorf("HourPath() = %q, want %q", got, want)
	}
}

func TestConceptHourPath(t *testing.T) {
	got := ConceptHourPath("acme", "2024", "my-hours")
	want := "acme/2024/concepts/hours/my-hours.toml"
	if got != want {
		t.Errorf("ConceptHourPath() = %q, want %q", got, want)
	}
}

func TestPurchasePath(t *testing.T) {
	tests := []struct {
		entity, year, bucket, name string
		paid                       bool
		want                       string
	}{
		{"acme", "2024", "Q1", "invoice-001", false, "acme/2024/Q1/purchase-invoices-unpaid/invoice-001.toml"},
		{"acme", "2024", "Q1", "invoice-001", true, "acme/2024/Q1/purchase-invoices-paid/invoice-001.toml"},
	}

	for _, tt := range tests {
		got := PurchasePath(tt.entity, tt.year, tt.bucket, tt.name, tt.paid)
		if got != tt.want {
			t.Errorf("PurchasePath(%q, %q, %q, %q, %v) = %q, want %q",
				tt.entity, tt.year, tt.bucket, tt.name, tt.paid, got, tt.want)
		}
	}
}

func TestValidateEntityPath(t *testing.T) {
	tests := []struct {
		path   string
		entity string
		ok     bool
	}{
		{"", "acme", true},                          // empty path is ok
		{"acme/2024/Q1/file.toml", "acme", true},    // valid path
		{"acme/subdir/file.toml", "acme", true},     // valid path
		{"other/2024/Q1/file.toml", "acme", false},  // wrong entity
		{"acme-corp/2024/file.toml", "acme", false}, // acme-corp != acme
		{"../acme/file.toml", "acme", false},        // traversal attempt
	}

	for _, tt := range tests {
		err := ValidateEntityPath(tt.path, tt.entity)
		if tt.ok && err != nil {
			t.Errorf("ValidateEntityPath(%q, %q) returned error: %v", tt.path, tt.entity, err)
		}
		if !tt.ok && err == nil {
			t.Errorf("ValidateEntityPath(%q, %q) should have returned error", tt.path, tt.entity)
		}
	}
}

func TestInvoiceSearchPaths(t *testing.T) {
	paths := InvoiceSearchPaths("acme", "2024", "Q1", "2024Q1-0001")
	if len(paths) != 2 {
		t.Errorf("InvoiceSearchPaths() returned %d paths, want 2", len(paths))
	}
	// First should be paid, second unpaid
	if paths[0] != "acme/2024/Q1/sales-invoices-paid/2024Q1-0001.toml" {
		t.Errorf("InvoiceSearchPaths()[0] = %q, want paid path", paths[0])
	}
	if paths[1] != "acme/2024/Q1/sales-invoices-unpaid/2024Q1-0001.toml" {
		t.Errorf("InvoiceSearchPaths()[1] = %q, want unpaid path", paths[1])
	}
}
