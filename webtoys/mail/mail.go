// Copyright 2017 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate bash packstatic.sh

package mail

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type MailReader struct {
	BoxPaths map[string]string
	Boxes    map[string]*Mailbox
}

type Header struct {
	Folder string
	Index  int

	From       string
	Subject    string
	PrettyDate string
	FullDate   string
	Date       time.Time `json:"-"`
}

type Message struct {
	MainHeaders  []string // html strings
	OtherHeaders []string // html strings
	Body         string   // html string
}

func (m *MailReader) folder(f string) (*Mailbox, error) {
	t0 := time.Now()
	fpath, ok := m.BoxPaths[f]
	if !ok {
		return nil, fmt.Errorf("no mail folder %q", f)
	}
	box := m.Boxes[f]
	if box != nil {
		return box, nil
	}

	fd, err := os.Open(fpath)
	if err != nil {
		return nil, fmt.Errorf("cannot open folder %s: %s", fpath, err)
	}
	box, err = Open(fd)
	if err != nil {
		fd.Close()
		return nil, fmt.Errorf("cannot open folder %s: %s", fpath, err)
	}
	m.Boxes[f] = box
	log.Printf("loaded mailbox %s in %.1fms", f,
		time.Since(t0).Seconds()*1000)
	return box, nil
}

func (m *MailReader) ListFolder(f string, start int) ([]Header, error) {
	now := time.Now()

	box, err := m.folder(f)
	if err != nil {
		return nil, err
	}

	msgs := box.msgs
	if start >= len(msgs) {
		return nil, nil
	}
	msgs = msgs[start:] // FIXME: pagination size

	var results []Header
	for i, m := range msgs {
		hdr := Header{
			Folder: f,
			Index:  start + i,

			From:       m.from,
			Subject:    m.subject,
			Date:       m.date,
			PrettyDate: prettyDate(now, m.date),
			FullDate:   m.date.Format("2006-01-02 15:04:05"),
		}
		results = append(results, hdr)
	}
	return results, nil
}

func prettyDate(now, t time.Time) string {
	t = t.In(now.Location())
	yy, mm, dd := now.Date()
	y, m, d := t.Date()
	switch {
	case yy == y && mm == m && dd == d:
		// same date
		return t.Format("today 15:04:05")
	case yy == y, // same year
		(y == yy-1 && m > mm): // last year
		return t.Format("02 Jan 15:04")
	default:
		return t.Format("02 Jan 2006")
	}
}

func (m *MailReader) Message(folder string, idx int) (*Message, error) {
	if idx < 0 {
		return nil, fmt.Errorf("negative index")
	}

	box, err := m.folder(folder)
	if err != nil {
		return nil, err
	}
	if idx >= len(box.msgs) {
		return nil, fmt.Errorf("no message with index %d", idx)
	}

	ms, err := box.Message(idx)
	if err != nil {
		return nil, err
	}

	mainHdrs := [...]string{"From", "Date", "Subject", "To", "Cc"}
	var msg Message
	for _, h := range mainHdrs {
		v := tryHeader(ms, h)
		if v != "" {
			msg.MainHeaders = append(msg.MainHeaders, h+": "+v)
		}
		delete(ms.Header, h)
	}
	for h := range ms.Header {
		v := tryHeader(ms, h)
		msg.OtherHeaders = append(msg.OtherHeaders, h+": "+v)
	}
	body, err := ioutil.ReadAll(ms.Body)
	if err != nil {
		return nil, err
	}
	msg.Body = string(body)
	return &msg, nil
}
