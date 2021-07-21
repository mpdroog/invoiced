// camt053 implements low-level XML parsing of the banktransaction
// file.
package camt053

import (
	"encoding/xml"
	"io"
)

type ParseStatement func(head GrpHdr, stmt Stmt) error

// Read token by token to easilly parse both big and small files :)
// http://blog.davidsingleton.org/parsing-huge-xml-files-with-go/
func Read(r io.Reader, fn ParseStatement) error {
	dec := xml.NewDecoder(r)

	for {
		t, e := dec.Token()
		if e != nil && e != io.EOF {
			return e
		}
		if t == nil {
			break
		}

		var head GrpHdr
		var stmt Stmt
		switch node := t.(type) {
		case xml.StartElement:
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
