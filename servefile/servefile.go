package main

import (
	"flag"
	"fmt"
	"net/http"
)

var (
	address, filename string
)

func main() {
	flag.StringVar(&address, "addr", "", "listen address")
	flag.StringVar(&filename, "file", "", "file name")
	flag.Parse()

	switch {
	case address == "", filename == "":
		flag.Usage()
	default:
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("connection from %s\n", r.RemoteAddr)
			http.ServeFile(w, r, filename)
		})
		http.ListenAndServe(address, nil)
	}
}
