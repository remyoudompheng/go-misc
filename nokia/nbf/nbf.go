// package nbf gives access to data from Nokia NBF archives.
package nbf

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"log"
	"path"
	"strings"
)

// OpenFile opens a NBF archive for reading.
func OpenFile(filename string) (*Reader, error) {
	z, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	return &Reader{z: z}, nil
}

type Reader struct {
	z *zip.ReadCloser
}

func (r *Reader) Close() error {
	return r.z.Close()
}

func (r *Reader) Inbox() ([]rawMessage, error) {
	msgs := make([]rawMessage, 0, len(r.z.File)/4)

	for _, f := range r.z.File {
		if !strings.HasPrefix(f.Name, "predefmessages/1/") {
			continue
		}
		base := path.Base(f.Name)
		//m.ParseFilename(base)
		//var m Message
		fr, err := f.Open()
		if err != nil {
			log.Printf("cannot read %s: %s", base, err)
			continue
		}
		blob, err := ioutil.ReadAll(fr)
		if err != nil {
			log.Printf("cannot read %s: %s", base, err)
			continue
		}
		m, err := parseMessage(blob)
		if err != nil {
			log.Printf("cannot parse %s: %s", base, err)
			continue
		}
		m.Filename = f.Name
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *Reader) Images() (images [][]byte, err error) {
	// convenience method to extract JPEG images
	for _, f := range r.z.File {
		if !strings.HasPrefix(f.Name, "predefmessages/") {
			continue
		}
		base := path.Base(f.Name)
		fr, err := f.Open()
		if err != nil {
			log.Printf("cannot read %s: %s", base, err)
			continue
		}
		blob, err := ioutil.ReadAll(fr)
		if err != nil {
			log.Printf("cannot read %s: %s", base, err)
			continue
		}
		count := 0
		for len(blob) > 0 {
			// look for 0xff 0xd8 ... JFIF ... 0xff 0xd9
			idx := bytes.Index(blob, []byte{0xff, 0xd8})
			if idx < 0 {
				break
			}
			idx0 := bytes.Index(blob[idx:], []byte{0xff, 0xd8})
			idx1 := bytes.Index(blob[idx:], []byte("JFIF"))
			idx2 := bytes.Index(blob[idx:], []byte{0xff, 0xd9})
			if idx0 > 0 && idx0 < idx1 {
				blob = blob[idx0:]
				continue
			} else if idx1 > 0 && idx2 > idx1 {
				count++
				log.Printf("found image %d in %s", count, base)
				images = append(images, blob[idx:idx+idx2+2])
				blob = blob[idx+idx2+2:]
			} else {
				break
			}
		}
	}
	return
}
