package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/remyoudompheng/go-misc/debug/go5"
	"github.com/remyoudompheng/go-misc/debug/go6"
	"github.com/remyoudompheng/go-misc/debug/go8"
	"github.com/remyoudompheng/go-misc/debug/goobj"
)

var cwd, _ = os.Getwd()

func cleanPath(s *string) {
	rel, err := filepath.Rel(cwd, *s)
	if err == nil && len(rel) < len(*s) {
		*s = rel
	}
}

func main() {
	obj := os.Args[1]
	f, err := os.Open(obj)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	first := true
	gochar := byte(0)
	for {
		line, err := rd.ReadSlice('\n')
		if err != nil && err != bufio.ErrBufferFull {
			log.Fatal(err)
		}
		if len(line) == 2 && string(line) == "!\n" {
			break
		}
		if first {
			first = false
			// go object GOOS GOARCH
			words := strings.Fields(string(line))
			arch := words[3]
			switch arch {
			case "arm":
				gochar = '5'
			case "amd64":
				gochar = '6'
			case "386":
				gochar = '8'
			}
		}
	}

	var fset *goobj.FileSet
	var imports map[int]string

	switch gochar {
	case '5':
		fset, imports = dump5(rd)
	case '6':
		fset, imports = dump6(rd)
	case '8':
		fset, imports = dump8(rd)
	default:
		log.Fatalf("unknown file type %s", obj)
	}

	fmt.Println("--- imports ---")
	for pos, imp := range imports {
		pos := fset.Position(pos)
		cleanPath(&pos.Filename)
		fmt.Printf("%s: imports %s\n", pos, imp)
	}
}

func dump5(r io.Reader) (fset *goobj.FileSet, imports map[int]string) {
	in := go5.NewReader(r)

	pcount := 0
	for {
		p, err := in.ReadProg()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		cleanPath(&p.Pos.Filename)
		switch p.Opname() {
		case "NAME", "HISTORY":
			// don't print.
			//fmt.Printf("%s\n", p)
		case "TEXT":
			fmt.Println()
			fmt.Printf("--- prog list %s ---\n", p.From.Sym)
			fallthrough
		default:
			fmt.Printf("%04d %s\n", pcount, p)
			pcount++
		case "END":
			break
		}
	}
	return in.Files()
}

func dump6(r io.Reader) (fset *goobj.FileSet, imports map[int]string) {
	in := go6.NewReader(r)

	pcount := 0
	for {
		p, err := in.ReadProg()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		cleanPath(&p.Pos.Filename)
		switch p.Opname() {
		case "NAME", "HISTORY":
			// don't print.
			//fmt.Printf("%s\n", p)
		case "TEXT":
			fmt.Println()
			fmt.Printf("--- prog list %s ---\n", p.From.Sym)
			fallthrough
		default:
			fmt.Printf("%04d %s\n", pcount, p)
			pcount++
		case "END":
			break
		}
	}
	return in.Files()
}

func dump8(r io.Reader) (fset *goobj.FileSet, imports map[int]string) {
	in := go8.NewReader(r)

	pcount := 0
	for {
		p, err := in.ReadProg()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		cleanPath(&p.Pos.Filename)
		switch p.Opname() {
		case "NAME", "HISTORY":
			// don't print.
			// fmt.Printf("%s\n", p)
		case "TEXT":
			fmt.Println()
			fmt.Printf("--- prog list %s ---\n", p.From.Sym)
			fallthrough
		default:
			fmt.Printf("%04d %s\n", pcount, p)
			pcount++
		case "END":
			break
		}
	}
	return in.Files()
}
