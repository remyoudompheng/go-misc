package main

import (
	"bufio"
	"fmt"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/remyoudompheng/go-misc/debug/go6"
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
	in := go6.NewReader(rd)

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

	fset, imports := in.Files()
	fmt.Println("--- imports ---")
	for pos, imp := range imports {
		pos := fset.Position(token.Pos(pos))
		pos.Line, pos.Column = pos.Column, 0
		cleanPath(&pos.Filename)
		fmt.Printf("%s: imports %s\n", pos, imp)
	}
}
