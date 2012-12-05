package pprof

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type HeapProfParser struct {
	R      *bufio.Reader
	lineno int
	Freq   int64
	// Total.
	LiveObj    int64
	LiveBytes  int64
	AllocObj   int64
	AllocBytes int64
}

var (
	heapRe           = regexp.MustCompile(`(\d+): (\d+) \[(\d+): (\d+)\] @`)
	errInvalidHeader = errors.New("malformed heap profile header")
)

type errBadLineHeader int

func (e errBadLineHeader) Error() string {
	return fmt.Sprintf("line %d: header in wrong format", int(e))
}

func isDigit(r rune) bool { return '0' <= r && r <= '9' }

// parseLine parses "xx: yyyy [zz: tttt] @" as the
// 4 numbers xx, yyyy, zz, tttt.
func parseLine(s []byte) (a, b, c, d int64, err error) {
	// s is in correct format if splitting at digits
	// yields ": ", " [", ": ", "] @".
	seps := bytes.FieldsFunc(s, isDigit)
	switch {
	case !bytes.Equal(seps[0], []byte(": ")),
		!bytes.Equal(seps[1], []byte(" [")),
		!bytes.Equal(seps[2], []byte(": ")),
		!bytes.Equal(seps[3], []byte("] @")):
		err = errBadLineHeader(0)
		return
	}
	for i, x := range s {
		if x == ':' {
			s = s[i+2:]
			break
		}
		a = a*10 + int64(x-'0')
	}
	for i, x := range s {
		if x == ' ' {
			s = s[i+2:]
			break
		}
		b = b*10 + int64(x-'0')
	}
	for i, x := range s {
		if x == ':' {
			s = s[i+2:]
			break
		}
		c = c*10 + int64(x-'0')
	}
	for i, x := range s {
		if x == ']' {
			s = s[i+2:]
			break
		}
		d = d*10 + int64(x-'0')
	}
	return
}

func NewHeapProfParser(r io.Reader) (*HeapProfParser, error) {
	const prefix = "heap profile: "
	b := bufio.NewReader(r)
	// Read totals.
	head, err := b.ReadSlice('@')
	if !bytes.HasPrefix(head, []byte(prefix)) {
		return nil, errInvalidHeader
	}
	head = head[len(prefix):]
	uo, ub, ao, ab, err := parseLine(head)
	if err != nil {
		println(string(head))
		return nil, err
	}
	p := &HeapProfParser{
		R:       b,
		LiveObj: uo, LiveBytes: ub,
		AllocObj: ao, AllocBytes: ab}
	p.lineno++
	// Read frequency.
	line, err := b.ReadSlice('\n')
	line = bytes.TrimSpace(line) // "heap/xxxx"
	if !bytes.HasPrefix(line, []byte("heap/")) {
		return nil, errInvalidHeader
	}
	line = line[5:]
	for _, x := range line {
		if x < '0' || x > '9' {
			return nil, errInvalidHeader
		}
		p.Freq = p.Freq*10 + int64(x-'0')
	}
	return p, nil
}

type HeapRecord struct {
	Trace []uint64 // A call trace (callee first).

	LiveObj    int64
	LiveBytes  int64
	AllocObj   int64
	AllocBytes int64
}

func (p *HeapProfParser) ReadRecord() (h HeapRecord, err error) {
	p.lineno++
	head, err := p.R.ReadSlice('@') // "xx: yy [zz: tt] @"
	if err != nil {
		return
	}
	lo, lb, ao, ab, err := parseLine(head)
	if err != nil {
		return h, errBadLineHeader(p.lineno)
	}
	h.LiveObj, h.LiveBytes = lo, lb
	h.AllocObj, h.AllocBytes = ao, ab
	// The call trace.
	line, err := p.R.ReadSlice('\n') // " 0x1234 0x2345 0x3456\n"
	line = bytes.TrimSpace(line)
	words := strings.Split(string(line), " ")
	trace := make([]uint64, len(words))
	for i, s := range words {
		// reverse stack trace.
		trace[i], err = strconv.ParseUint(s, 0, 64) // parse 0x1234
		if err != nil {
			return
		}
	}
	h.Trace = trace
	return h, nil
}
