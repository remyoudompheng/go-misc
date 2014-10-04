// package nbf gives access to data from Nokia NBF archives.
package nbf

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"log"
	"path"
	"sort"
	"strings"
	"time"
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

type SMS struct {
	Type int // 0: incoming, 1: outgoing
	Peer string
	When time.Time
	Text string
}

func (r *Reader) Inbox() ([]SMS, error) {
	msgs := make([]SMS, 0, len(r.z.File)/4)

	type multiKey struct {
		Peer string
		Ref  int
	}
	multiparts := make(map[multiKey][]deliverMessage)

	for _, f := range r.z.File {
		if !strings.HasPrefix(f.Name, "predefmessages/1/") {
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
		m, err := parseMessage(blob)
		if err != nil {
			log.Printf("cannot parse %s: %s", base, err)
			continue
		}

		sms := SMS{
			Type: int(m.Msg.MsgType),
			Peer: m.Msg.FromAddr,
			When: m.Msg.SMSCStamp,
			Text: m.Msg.UserData(),
		}

		if m.Msg.Concat {
			key := multiKey{Peer: sms.Peer, Ref: m.Msg.Ref}
			parts := append(multiparts[key], m.Msg)
			if len(parts) == m.Msg.NParts {
				delete(multiparts, key)
				p := make(map[int]string)
				for _, part := range parts {
					p[part.Part] = part.UserData()
				}
				sms.Text = ""
				for i := 1; i <= m.Msg.NParts; i++ {
					sms.Text += p[i]
				}
				msgs = append(msgs, sms)
			} else {
				multiparts[key] = parts
			}
		} else {
			msgs = append(msgs, sms)
		}
	}
	sort.Sort(smsByDate(msgs))
	return msgs, nil
}

type smsByDate []SMS

func (s smsByDate) Len() int           { return len(s) }
func (s smsByDate) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s smsByDate) Less(i, j int) bool { return s[i].When.Before(s[j].When) }

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
				//log.Printf("found image %d in %s", count, base)
				images = append(images, blob[idx:idx+idx2+2])
				blob = blob[idx+idx2+2:]
			} else {
				break
			}
		}
	}
	return
}
