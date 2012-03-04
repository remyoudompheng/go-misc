package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/remyoudompheng/go-misc/webtoys/pastehere"
	_ "github.com/remyoudompheng/go-misc/webtoys/webclock"
)

var logger = log.New(os.Stderr, "webtoys ", log.LstdFlags)

var toys = []string{
	"pastehere",
	"webclock",
}

const indexTplString = `<!DOCTYPE html>
<html>
<head>
	<title>Rémy's Webtoys</title>
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
</head>
<body>
	<h1>Rémy's Webtoys</h1>

	<ul>
	{{ range $ }}
	<li><a href="/{{ . }}">{{ . }}</a></li>
	{{ end }}
	</ul>
</body>
</html>
`

var indexTpl = template.Must(template.New("index").Parse(indexTplString))

func index(resp http.ResponseWriter, req *http.Request) {
	logger.Printf("GET %s from %s", req.URL, req.RemoteAddr)
	if req.URL.Path != "/" {
		resp.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(resp, "No such page: %s", req.URL.Path)
		return
	}
	indexTpl.Execute(resp, toys)
}

func init() {
	http.HandleFunc("/", index)
}

func main() {
	var address string
	flag.StringVar(&address, "http", ":8080", "listen address")
	flag.Parse()
	if address == "" {
		flag.Usage()
		return
	}
	logger.Printf("start listening at %s", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		logger.Fatal(err)
	}
}
