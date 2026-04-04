package db

import (
	"fmt"
	"strings"
)

// CommitAction represents the type of action being performed.
type CommitAction string

// Common commit actions following conventional commits style.
const (
	ActionCreate   CommitAction = "create"
	ActionUpdate   CommitAction = "update"
	ActionDelete   CommitAction = "delete"
	ActionFinalize CommitAction = "finalize"
	ActionPay      CommitAction = "pay"
	ActionUnpay    CommitAction = "unpay"
	ActionEmail    CommitAction = "email"
	ActionBill     CommitAction = "bill"
	ActionImport   CommitAction = "import"
)

// CommitResource represents the type of resource being modified.
type CommitResource string

// Resource types for commit messages.
const (
	ResourceInvoice         CommitResource = "invoice"
	ResourcePurchaseInvoice CommitResource = "purchase"
	ResourceHour            CommitResource = "hour"
	ResourceDebtor          CommitResource = "debtor"
	ResourceProject         CommitResource = "project"
	ResourceBank            CommitResource = "bank"
)

// FormatCommitMsg creates a standardized commit message.
// Format: [entity] action(resource): id details
// Example: [acme] finalize(invoice): 2024-0042 €1,250.00
func FormatCommitMsg(entity string, action CommitAction, resource CommitResource, id string, details ...string) string {
	var msg strings.Builder

	fmt.Fprintf(&msg, "[%s] %s(%s): %s", entity, action, resource, id)

	if len(details) > 0 {
		msg.WriteString(" ")
		msg.WriteString(strings.Join(details, " "))
	}

	return msg.String()
}

// FormatAmount formats a monetary amount for commit messages.
func FormatAmount(currency string, amount float64) string {
	if currency == "" {
		currency = "EUR"
	}
	symbol := "€"
	switch currency {
	case "USD":
		symbol = "$"
	case "GBP":
		symbol = "£"
	}
	return fmt.Sprintf("%s%.2f", symbol, amount)
}
