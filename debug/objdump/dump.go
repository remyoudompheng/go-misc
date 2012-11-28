package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/remyoudompheng/go-misc/debug/go6"
)

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

	for {
		p, err := in.ReadProg()
		if err != nil {
			log.Fatal(err)
		}
		if p.Op == go6.ATEXT {
			fmt.Println()
		}
		fmt.Println(p)
		if p.Op == go6.AEND {
			return
		}
	}
}
