// nbfextract is a utility that dumps contents of a NBF archive
// into mainstream format files.
package main

import (
	"fmt"
	"log"
	"os"

	//"github.com/remyoudompheng/go-misc/nokia/mms"
	"github.com/remyoudompheng/go-misc/nokia/nbf"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s input.nbf destdir/", os.Args[0])
		os.Exit(1)
	}
	input := os.Args[1]
	destdir := os.Args[2]

	log.Printf("dumping %s to %s", input, destdir)
	f, err := nbf.OpenFile(input)
	if err != nil {
		log.Fatalf("could not open %s: %s", input, err)
	}
	defer f.Close()

	inbox, err := f.Inbox()
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range inbox {
		//log.Print(m.Filename)
		text := m.Text
		if text == "" {
			text = m.Msg.UserData()
		}
		log.Printf("%s at %s : %q", m.Peer, m.Msg.SMSCStamp, text)
	}
}
