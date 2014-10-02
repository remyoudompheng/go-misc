// package nbf gives access to data from Nokia NBF archives.
package nbf

import (
	"archive/zip"
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
		}
		blob, err := ioutil.ReadAll(fr)
		if err != nil {
			log.Printf("cannot read %s: %s", base, err)
		}
		m, err := parseMessage(blob)
		if err != nil {
			log.Printf("cannot parse %s: %s", base, err)
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}
