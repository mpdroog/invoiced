// UBL XML Invoice
// https://www.jortt.nl/online-boekhouden/ubl-factuur/
package invoice

import (
	"bytes"
	"text/template"
	//"encoding/xml"
	"github.com/mpdroog/invoiced/embed"
)

func UBL(u *Invoice) (*bytes.Buffer, error) {
	raw := embed.MustAsset("embed/invoice.xml")
	tpl := template.New("UBL")
	//tpl.Funcs(map[string]interface{} {
	//	"escape": xml.EscapeText,
	//})

	tpl, e := tpl.Parse(string(raw))
	if e != nil {
		return nil, e
	}

	b := new(bytes.Buffer)
	e = tpl.Execute(b, u)
	return b, e
}