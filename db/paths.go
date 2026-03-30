package db

import (
	"fmt"
	"strings"
)

// InvoicePath returns the path for an invoice file.
// bucket should be "Q1", "Q2", "Q3", or "Q4".
// paid indicates whether this is a paid or unpaid invoice.
func InvoicePath(entity, year, bucket, name string, paid bool) string {
	status := "sales-invoices-unpaid"
	if paid {
		status = "sales-invoices-paid"
	}
	return fmt.Sprintf("%s/%s/%s/%s/%s.toml", entity, year, bucket, status, name)
}

// ConceptInvoicePath returns the path for a concept invoice.
func ConceptInvoicePath(entity, year, name string) string {
	return fmt.Sprintf("%s/%s/concepts/sales-invoices/%s.toml", entity, year, name)
}

// HourPath returns the path for an hour registration file.
// bucket should be "Q1", "Q2", "Q3", or "Q4".
func HourPath(entity, year, bucket, name string) string {
	return fmt.Sprintf("%s/%s/%s/hours/%s.toml", entity, year, bucket, name)
}

// ConceptHourPath returns the path for a concept hour registration.
func ConceptHourPath(entity, year, name string) string {
	return fmt.Sprintf("%s/%s/concepts/hours/%s.toml", entity, year, name)
}

// PurchasePath returns the path for a purchase invoice file.
// paid indicates whether this is a paid or unpaid purchase invoice.
func PurchasePath(entity, year, bucket, name string, paid bool) string {
	status := "purchase-invoices-unpaid"
	if paid {
		status = "purchase-invoices-paid"
	}
	return fmt.Sprintf("%s/%s/%s/%s/%s.toml", entity, year, bucket, status, name)
}

// DebtorsPath returns the path for the debtors file.
func DebtorsPath(entity string) string {
	return fmt.Sprintf("%s/debtors.toml", entity)
}

// ProjectsPath returns the path for the projects file.
func ProjectsPath(entity string) string {
	return fmt.Sprintf("%s/projects.toml", entity)
}

// LogoPath returns the path for the entity logo.
func LogoPath(entity string) string {
	return fmt.Sprintf("%s/logo.png", entity)
}

// ValidateEntityPath checks that a path belongs to the expected entity.
// Returns an error if the path doesn't start with the entity prefix.
func ValidateEntityPath(path, entity string) error {
	if path == "" {
		return nil
	}
	if !strings.HasPrefix(path, entity+"/") {
		return fmt.Errorf("invalid path: must belong to entity %s", entity)
	}
	return nil
}

// InvoiceSearchPaths returns the paths to search for an invoice (both paid and unpaid).
func InvoiceSearchPaths(entity, year, bucket, name string) []string {
	return []string{
		InvoicePath(entity, year, bucket, name, true),
		InvoicePath(entity, year, bucket, name, false),
	}
}

// InvoiceListPaths returns the glob paths for listing all invoices.
func InvoiceListPaths(entity, year string) []string {
	return []string{
		fmt.Sprintf("%s/%s/concepts/sales-invoices", entity, year),
		fmt.Sprintf("%s/%s/{all}/sales-invoices-paid", entity, year),
		fmt.Sprintf("%s/%s/{all}/sales-invoices-unpaid", entity, year),
	}
}

// PurchaseListPaths returns the glob paths for listing all purchase invoices.
func PurchaseListPaths(entity, year string) []string {
	return []string{
		fmt.Sprintf("%s/%s/{all}/purchase-invoices-unpaid", entity, year),
		fmt.Sprintf("%s/%s/{all}/purchase-invoices-paid", entity, year),
	}
}

// HourListPaths returns the glob paths for listing all hours.
func HourListPaths(entity, year string) []string {
	return []string{
		fmt.Sprintf("%s/%s/{all}/hours", entity, year),
	}
}
