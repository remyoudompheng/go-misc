package main

import (
	"io"
	"log"
	"os"
	"time"
)

func main() {
	files := os.Args[1:]
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			log.Printf("%s: %s", file, err)
			continue
		}
		for {
			box, err := ReadBox(f)
			if err == io.EOF {
				f.Close()
				break
			}
			if err != nil {
				f.Close()
				log.Printf("%s: %s", file, err)
				break
			}
			log.Printf("%s: box %s (%d bytes)", file, box.Type, len(box.Data))
			handleBox(box)
		}
	}
}

var wroteHdr = false

var timeScale uint32

func handleBox(box Box) {
	var err error
	switch box.Type {
	default:
		return
	case "abst":
		var binfo BootstrapInfo
		binfo, err = handleBootstrapInfo(box)
		timeScale = binfo.TimeScale
	case "mdat":
		frames := handleMovieData(box)
		for _, f := range frames {
			if f.IsSeqHeader() {
				log.Printf("skipping %s (%d bytes)", f.Describe(), len(f.Data))
				continue
			}
			stamp := time.Second * time.Duration(f.Stamp) / time.Duration(timeScale)
			log.Printf("frame at %s: %s (%d bytes)", stamp, f.Describe(), len(f.Data))
			if !wroteHdr {
				wroteHdr = true
				err = writeFLVHeader(os.Stdout)
				_, err = os.Stdout.Write(box.Data)
			}
			if err == nil {
				err = f.WriteTo(os.Stdout)
			}
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	if err != nil {
		log.Printf("error in box %s: %s", box.Type, err)
	}
}

func writeFLVHeader(w io.Writer) error {
	// See E.2 The FLV Header
	_, err := io.WriteString(w, "FLV\x01")
	if err != nil {
		return err
	}
	_, err = w.Write([]byte{
		1<<2 | 1,   // Audio/Video
		0, 0, 0, 9, // length of header
		0, 0, 0, 0, // PreviousTagSize0 (E.3)
	})
	return err
}
