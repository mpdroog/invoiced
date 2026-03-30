// Package camt053 implements low-level XML parsing of the banktransaction
// file.
package camt053

import (
	"encoding/xml"
	"errors"
	"io"
)

// ParseStatement is a callback function for processing bank statements.
type ParseStatement func(head GrpHdr, stmt Stmt) error

// Read parses CAMT053 XML token by token to handle both big and small files.
func Read(r io.Reader, fn ParseStatement) error {
	dec := xml.NewDecoder(r)

	for {
		t, e := dec.Token()
		if e != nil && !errors.Is(e, io.EOF) {
			return e
		}
		if t == nil {
			break
		}

		var head GrpHdr
		var stmt Stmt
		if node, ok := t.(xml.StartElement); ok {
			if node.Name.Local == "GrpHdr" {
				if e := dec.DecodeElement(&head, &node); e != nil {
					return e
				}
			}
			if node.Name.Local == "Stmt" {
				if e := dec.DecodeElement(&stmt, &node); e != nil {
					return e
				}
				if e := fn(head, stmt); e != nil {
					return e
				}
			}
		}
	}

	return nil
}
