package vdeck

import (
	"bytes"
	"encoding/json"
	"flag"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var logger = log.New(os.Stderr, "vdeck ", log.LstdFlags|log.Lshortfile)

func init() {
	flag.StringVar(&vcardDir, "vdeck", "", "vCard directory path")
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
		vc.Filename, _ = filepath.Rel(dirname, path)

		cards = append(cards, vc)
		return nil
	})
	return cards
}

func loadCard(name string) (*VCard, error) {
	fname := filepath.Join(vcardDir, name)
	if !strings.HasPrefix(fname, vcardDir) {
		return nil, fmt.Errorf("wrong path %s", fname)
	}
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseVcard(f)
}

const indexTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <title>vCard explorer</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <link type="text/css" rel="stylesheet" href="/static/jqgrid/ui.jqgrid.css"/>
    <script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.4.2/jquery.min.js"></script>
    <script type="text/javascript" src="/static/jqgrid/jquery.jqGrid.min.js"></script>
    <script type="text/javascript" src="/static/vdeck.js"></script>
  </head>
  <body>
    <h1>Contacts directory</h1>

    <table id="contacts"></table>
  </body>
</html>
`

var indexTpl = template.Must(template.
	New("index").
	Parse(indexTemplate))

func init() {
	http.HandleFunc("/vdeck/", index)
	http.HandleFunc("/vdeck/all/", index_jqgrid)
	http.HandleFunc("/vdeck/vcf/", raw_vcard)
	logger.Printf("registered vdeck at /vdeck/")
}

func index(w http.ResponseWriter, req *http.Request) {
	cards := loadDirectory(vcardDir)
	indexTpl.Execute(w, cards)
}

func index_jqgrid(w http.ResponseWriter, req *http.Request) {
	type record struct {
		FullName   string `json:"fullname"`
		FamilyName string `json:"family_name"`
		FirstName  string `json:"first_name"`
		Phone      string `json:"phone"`
		Email      string `json:"email"`
		Filename   string `json:"filename"`
		Uid        string `json:"uid"`
	}

	type records struct {
		Total   int      `json:"total"`
		Page    int      `json:"page"`
		Records int      `json:"records"`
		Rows    []record `json:"rows"`
	}

	w.Header().Set("Content-Type", "application/json")
	cards := loadDirectory(vcardDir)
	data := records{
		Total: 1, Page: 1,
		Records: len(cards),
		Rows:    make([]record, len(cards)),
	}
	for i, c := range cards {
		data.Rows[i] = record{
			FullName:   c.FullName,
			FamilyName: c.Name.FamilyName,
			FirstName:  c.Name.GivenName,
			Filename:   c.Filename,
			Uid:        c.Uid,
		}
		if len(c.Tel) > 0 {
			data.Rows[i].Phone = c.Tel[0].Value
		}
		if len(c.Email) > 0 {
			data.Rows[i].Email = c.Email[0].Value
		}
	}
	json.NewEncoder(w).Encode(data)
}

func raw_vcard(w http.ResponseWriter, req *http.Request) {
	cardpath, err := filepath.Rel("/vdeck/vcf/", req.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if card, err := loadCard(cardpath); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		w.Header().Set("Content-Type", "text/x-vcard")
		io.WriteString(w, card.String())
	}
}
