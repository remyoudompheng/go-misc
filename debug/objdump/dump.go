package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/remyoudompheng/go-misc/debug/ar"
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

	// Read first line.
	line, err := rd.Peek(8)
	if err != nil {
		log.Fatal(err)
	}
	switch string(line) {
	case "!<arch>\n":
		dumparchive(rd)
	case "go objec":
		dumpobj(rd)
	default:
		log.Fatalf("unknown file type %s: bad magic %q", obj, line)
	}
}

func dumpobj(rd *bufio.Reader) {
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
			default:
				log.Printf("unrecognized object format %s", line)
				return
			}
		}
	}

	switch gochar {
	case '5':
		dump(Reader5{go5.NewReader(rd)})
	case '6':
		dump(Reader6{go6.NewReader(rd)})
	case '8':
		dump(Reader8{go8.NewReader(rd)})
	}
}

func dumparchive(rd *bufio.Reader) {
	r := ar.NewReader(rd)
	for {
		hdr, err := r.Next()
		switch err {
		case nil:
		case io.EOF:
			return
		default:
			log.Fatal(err)
		}
		switch hdr.Name {
		case "__.PKGDEF", "__.GOSYMDEF":
			continue
		default:
			fmt.Printf("--- object %s ---\n", hdr.Name)
			dumpobj(bufio.NewReader(r))
		}
	}
}

func dump(r ProgReader) {
	pcount := 0
	for {
		p, pos, sym, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		cleanPath(&pos.Filename)
		switch p.Opname() {
		case "NAME", "HISTORY":
			// don't print.
			//fmt.Printf("%s\n", p)
		case "TEXT":
			fmt.Println()
			fmt.Printf("--- prog list %s ---\n", sym)
			fallthrough
		default:
			fmt.Printf("%04d %s\n", pcount, p)
			pcount++
		case "END":
			break
		}
	}

	fset, imports := r.Files()
	fmt.Println("--- imports ---")
	for pos, imp := range imports {
		pos := fset.Position(pos)
		cleanPath(&pos.Filename)
		fmt.Printf("%s: imports %s\n", pos, imp)
	}
}

type Prog interface {
	Opname() string
}

type ProgReader interface {
	Read() (Prog, *goobj.Position, string, error)
	Files() (*goobj.FileSet, map[int]string)
}

type Reader5 struct{ *go5.Reader }

func (r Reader5) Read() (Prog, *goobj.Position, string, error) {
	prog, err := r.Reader.ReadProg()
	return &prog, &prog.Pos, prog.From.Sym, err
}

type Reader6 struct{ *go6.Reader }

func (r Reader6) Read() (Prog, *goobj.Position, string, error) {
	prog, err := r.Reader.ReadProg()
	return &prog, &prog.Pos, prog.From.Sym, err
}

type Reader8 struct{ *go8.Reader }

func (r Reader8) Read() (Prog, *goobj.Position, string, error) {
	prog, err := r.Reader.ReadProg()
	return &prog, &prog.Pos, prog.From.Sym, err
}
