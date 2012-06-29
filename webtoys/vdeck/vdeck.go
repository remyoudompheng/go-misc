package vdeck

import (
	"bytes"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var logger = log.New(os.Stderr, "vdeck ", log.LstdFlags|log.Lshortfile)

func init() {
	flag.StringVar(&vcardDir, "vdeck", "", "vCard directory path")
	http.HandleFunc("/vdeck/", index)
	logger.Printf("registered vdeck at /vdeck/")
}

var vcardDir string // the base directory for vCards.

func loadDirectory(dirname string) []*VCard {
	cards := make([]*VCard, 0, 64)
	var errors []error
	filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".vcf") {
			return nil
		}
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			errors = append(errors, err)
			return nil
		}

		vc, err := ParseVcard(bytes.NewReader(contents))
		if err != nil {
			errors = append(errors, err)
			return nil
		}
		vc.Filename = path

		cards = append(cards, vc)
		return nil
	})
	return cards
}

const indexTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <title>vCard explorer</title>
  </head>
  <body>
    <h1>Contacts directory</h1>

    <table>
    <thead>
      <tr>
	  <th>Full name</th>
	  <th>Family name</th>
	  <th>First name</th>
	  <th>Phone number</th>
	  <th>Email</th>
	  <th>Filename</th>
	</tr>
    </thead>
    <tbody>
      {{ range $card := . }}
	<tr>
	  <td>{{ $card.FullName }}</td>
	  <td>{{ $card.Name.FamilyName }}</td>
	  <td>{{ $card.Name.GivenName }}</td>
	  <td>{{ head $card.Tel }}</td>
	  <td>{{ head $card.Email }}</td>
	  <td>{{ $card.Filename }}</td>
	</tr>
	{{ end }}
    </tbody>
    </table>
  </body>
</html>
`

var indexTpl = template.Must(template.
	New("index").
	Funcs(template.FuncMap{"head": tpl_head}).
	Parse(indexTemplate))

func tpl_head(slice interface{}) interface{} {
	v := reflect.ValueOf(slice)
	if v.Len() == 0 {
		return ""
	}
	return v.Index(0).Interface()
}

func index(w http.ResponseWriter, req *http.Request) {
	cards := loadDirectory(vcardDir)
	indexTpl.Execute(w, cards)
}
