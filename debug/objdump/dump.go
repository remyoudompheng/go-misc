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
	if err == nil {
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
	for {
		line, err := rd.ReadSlice('\n')
		if err != nil && err != bufio.ErrBufferFull {
			log.Fatal(err)
		}
		if len(line) == 2 && string(line) == "!\n" {
			break
		}
	}

	var fset *goobj.FileSet
	var imports map[int]string

	switch {
	case strings.HasSuffix(obj, ".5"):
		fset, imports = dump5(rd)
	case strings.HasSuffix(obj, ".6"):
		fset, imports = dump6(rd)
	case strings.HasSuffix(obj, ".8"):
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
		switch p.Op {
		case go5.ANAME, go5.AHISTORY:
			// don't print.
			//fmt.Printf("%s\n", p)
		case go5.ATEXT:
			fmt.Println()
			fmt.Printf("--- prog list %s ---\n", p.From.Sym)
			fallthrough
		default:
			fmt.Printf("%04d %s\n", pcount, p)
			pcount++
		case go5.AEND:
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
		switch p.Op {
		case go6.ANAME, go6.AHISTORY:
			// don't print.
			//fmt.Printf("%s\n", p)
		case go6.ATEXT:
			fmt.Println()
			fmt.Printf("--- prog list %s ---\n", p.From.Sym)
			fallthrough
		default:
			fmt.Printf("%04d %s\n", pcount, p)
			pcount++
		case go6.AEND:
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
		switch p.Op {
		case go8.ANAME, go8.AHISTORY:
			// don't print.
			//fmt.Printf("%s\n", p)
		case go8.ATEXT:
			fmt.Println()
			fmt.Printf("--- prog list %s ---\n", p.From.Sym)
			fallthrough
		default:
			fmt.Printf("%04d %s\n", pcount, p)
			pcount++
		case go8.AEND:
			break
		}
	}
	return in.Files()
}
