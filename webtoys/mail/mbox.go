// Copyright 2017 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mail

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/mail"
)

// This file implements access to mbox files.
// The mbox source file is expected to be a random access
// SectionReader, which will be indexed at first access.
//
// The implementation assumes quoting of lines starting with "From "
// aka the "mboxrd" format.

type Mailbox struct {
	r    io.ReaderAt
	msgs []mboxMsg
}

type mboxMsg struct {
	offset  int
	length  int
	from    string
	subject string
	date    string
}

const maxMessageSize = 50 * 1024 * 1024 // 50MB should be enough for everybody

func Open(r io.ReaderAt) (*Mailbox, error) {
	// r is usually a reader, for example if it is *os.File
	rd, ok := r.(io.Reader)
	if !ok {
		// 1GB should be enough for everybody :)
		rd = io.NewSectionReader(r, 0, 1<<30)
	}
	s := bufio.NewScanner(rd)
	s.Buffer(nil, maxMessageSize)
	s.Split(scanMessage)

	offset := 0
	var msgs []mboxMsg
	for s.Scan() {
		data := s.Bytes()
		msg := mboxMsg{offset: offset, length: len(data)}
		m, err := mail.ReadMessage(bytes.NewReader(data))
		if err == nil {
			msg.from = tryHeader(m, "From")
			msg.subject = tryHeader(m, "Subject")
			msg.date = tryHeader(m, "Date")
		}
		msgs = append(msgs, msg)
		offset += len(data)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	box := &Mailbox{
		r:    r,
		msgs: msgs,
	}
	return box, nil
}

func scanMessage(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 && atEOF {
		return 0, nil, nil
	}
	if !bytes.HasPrefix(data, []byte("From ")) {
		return 0, nil, fmt.Errorf("invalid mbox file: first line %30q", data)
	}
	end := bytes.Index(data, []byte("\n\nFrom "))
	if end >= 0 {
		return end + 2, data[:end+2], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func (m *Mailbox) Message(idx int) (*mail.Message, error) {
	msg := m.msgs[idx]
	r := io.NewSectionReader(m.r, int64(msg.offset), int64(msg.length))
	buf := bufio.NewReader(r)
	buf.ReadBytes('\n') // discard first line
	return mail.ReadMessage(buf)
}

func tryHeader(msg *mail.Message, hdr string) string {
	v := msg.Header.Get(hdr)
	dec := mime.WordDecoder{CharsetReader: charsetReader}
	s, err := dec.DecodeHeader(v)
	if err == nil {
		return s
	}
	return v
}

// Additional encodings

func charsetReader(encoding string, r io.Reader) (io.Reader, error) {
	switch encoding {
	case "iso-8859-15":
		s, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return decodeLatin9(s), nil
	case "cp1252", "windows-1252", "windows-1258":
		// FIXME: Windows-1258 is NOT equivalent to Windows-1252
		// but in our case there is no practical difference
		s, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return decodeWin1252(s), nil
	case "euc-kr":
		// FIXME: use latin9 as a hack to at least recognize the ASCII
		// subset
		s, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return decodeLatin9(s), nil
	default:
		return nil, fmt.Errorf("unsupported encoding %s", encoding)
	}
}

// decodeLatin9 decodes ISO-8859-15
func decodeLatin9(s []byte) *bytes.Buffer {
	buf := new(bytes.Buffer)
	for _, b := range s {
		char := rune(b)
		// ISO-8859-15 is the same as ISO-8859-1 (first Unicode code points)
		// except for a few characters.
		switch b {
		case 0xa4:
			char = '€'
		case 0xa6:
			char = 'Š'
		case 0xa8:
			char = 'š'
		case 0xb4:
			char = 'Ž'
		case 0xb8:
			char = 'ž'
		case 0xbc:
			char = 'Œ'
		case 0xbd:
			char = 'œ'
		case 0xbe:
			char = 'Ÿ'
		}
		buf.WriteRune(char)
	}
	//println("LATIN9", string(s), buf.String())
	return buf
}

var win1252special = [...]rune{
	// \x80 .. \x8f
	'€', '\ufffd', '‚', 'ƒ', '„', '…', '†', '‡', 'ˆ', '‰', 'Š', '‹', 'Œ', '\ufffd', 'Ž', '\ufffd',
	// \x90 .. \x9f
	'\ufffd', '‘', '’', '“', '”', '•', '–', '—', '˜', '™', 'š', '›', 'œ', '\ufffd', 'ž', 'Ÿ',
}

func decodeWin1252(s []byte) *bytes.Buffer {
	buf := new(bytes.Buffer)
	for _, b := range s {
		char := rune(b)
		// Windows-1252 looks like ISO-8859-1 except for the
		// bytes in the 0x80-0x9F range.
		if 0x80 <= b && b < 0xa0 {
			char = win1252special[b-0x80]
		}
		buf.WriteRune(char)
	}
	//println("WIN1252", buf.String())
	return buf
}
