// package weechat implements the WeeChat relay protocol.
package weechat

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"reflect"
	"strconv"
      "time"
)

// Reference: http://www.weechat.org/files/doc/stable/weechat_relay_protocol.en.html

type command int

const (
	cmdInit command = iota
	cmdHdata
	cmdInfo
	cmdInfolist
	cmdNicklist
	cmdInput
	cmdSync
	cmdDesync
	cmdQuit
	cmdCount
)

var cmdStrings = [cmdCount]string{
	cmdInit:     "init",
	cmdHdata:    "hdata",
	cmdInfo:     "info",
	cmdInfolist: "infolist",
	cmdNicklist: "nicklist",
	cmdInput:    "input",
	cmdSync:     "sync",
	cmdDesync:   "desync",
	cmdQuit:     "quit",
}

type Conn struct {
	c net.Conn
	r *bufio.Reader
}

func Dial(addr string) (*Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Conn{c: conn, r: bufio.NewReader(conn)}, nil
}

func (conn *Conn) send(cmd command, args ...string) error {
	buf := make([]byte, 0, 80)
	buf = append(buf, cmdStrings[cmd]...)
	for _, a := range args {
		buf = append(buf, ' ')
		buf = append(buf, a...)
	}
	buf = append(buf, '\n')
	_, err := conn.c.Write(buf)
	return err
}

var errMsgTooLarge = errors.New("message too large")

// recv gets a message from the connection.
func (conn *Conn) recv() (s []byte, err error) {
	// A message is:
	// - a uint32 length
	// - a byte boolean for compression
	// - length-5 bytes of data (plain or zlib compressed)
	var buf [5]byte
	_, err = io.ReadFull(conn.r, buf[:])
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(buf[:4])
	isCompressed := buf[4] == 1
	if length >= 32<<20 {
		return nil, errMsgTooLarge
	}

	s = make([]byte, length-5)
	_, err = io.ReadFull(conn.r, s)
	if err != nil {
		return
	}
	if isCompressed {
		zr, err := zlib.NewReader(bytes.NewBuffer(s))
		if err != nil {
			return s, err
		}
		return ioutil.ReadAll(zr)
	}
	return s, nil
}

const (
	typeChar   = "chr"
	typeInt    = "int"
	typeLong   = "lon"
	typeString = "str"
	typeBytes  = "buf"
	typePtr    = "ptr"
	typeTime   = "tim"
	typeMap    = "htb"
	typeHdata  = "hda"
	typeInfo   = "inf"
	typeInfos  = "inl"
	typeArray  = "arr"
)

var types = []string{
	typeChar, typeInt, typeLong,
	typeString, typeBytes,
	typePtr, typeTime,
	typeMap, typeHdata, typeArray,
	typeInfo, typeInfos,
}

type message []byte

func (m *message) Int() int32 {
	u := binary.BigEndian.Uint32((*m)[:4])
	*m = (*m)[4:]
	return int32(u)
}

func (m *message) Byte() byte {
	b := (*m)[0]
	*m = (*m)[1:]
	return b
}

func (m *message) Pointer() uint64 {
	l := m.Byte()
	p, err := strconv.ParseUint(string((*m)[:l]), 16, 64)
	if err != nil {
		panic(err)
	}
	*m = (*m)[l:]
	return p
}

func (m *message) Buffer() []byte {
	length := m.Int()
	if length == -1 {
		return nil
	}
	data := (*m)[:length]
	*m = (*m)[length:]
	return data
}

func (m *message) Time() time.Time {
	// a 1-byte length + base 10 integer (Unix)
	length := int(m.Byte())
	t, err := strconv.ParseInt(string((*m)[:length]), 10, 64)
	if err != nil {
		panic(err)
	}
	*m = (*m)[length:]
	return time.Unix(t, 0)
}

func (m *message) GetType() string {
	var buf [3]byte
	copy(buf[:], (*m)[:3])
	*m = (*m)[3:]
	for _, t := range types {
		if buf[0] == t[0] && buf[1] == t[1] && buf[2] == t[2] {
			return t
		}
	}
	panic("invalid type " + string(buf[:]))
}

// v is a settable slice.
func (m *message) Hdata(v reflect.Value) (ppaths [][]uint64) {
	if v.Kind() != reflect.Slice || !v.CanSet() {
		panic("not a settable slice")
	}
	hpath := bytes.Split(m.Buffer(), []byte{'/'}) // "buffer/nick_group"
	keys := bytes.Split(m.Buffer(), []byte{','})  // "name:str,prefix:str"a
	// make result slice
	length := m.Int()
	slice := reflect.MakeSlice(v.Type(), int(length), int(length))
	v.Set(slice)
	// map key names to field index.
	keytofield := make([]int, len(keys))
	t := v.Type().Elem()
	for i := range keys {
		keytofield[i] = -1
	}
loopfields:
	for f, fcount := 0, t.NumField(); f < fcount; f++ {
		fld := t.Field(f)
		for k, key := range keys {
			if cmpbytestring(key[:len(key)-4], string(fld.Tag)) {
				keytofield[k] = f
				continue loopfields
			}
		}
	}
	for i := int32(0); i < length; i++ {
		// len(hpath) pointers
		// len(keys) objects with the given types.
		ppath := make([]uint64, len(hpath))
		for p := range hpath {
			ppath[p] = m.Pointer()
		}
		obj := v.Index(int(i))
		for k, fld := range keytofield {
			km := message(keys[k])
			km = km[len(km)-3:]
			keytype := km.GetType()
			var dst reflect.Value
			if fld >= 0 {
				dst = obj.Field(fld)
			} else if i == 0 {
				debugf("key %s ignored in %s", keys[k], hpath)
			}
			m.decodeValue(keytype, dst)
		}
		ppaths = append(ppaths, ppath)
	}
	return ppaths
}

func cmpbytestring(a []byte, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, c := range a {
		if c != b[i] {
			return false
		}
	}
	return true
}

func (m *message) decodeValue(typ string, v reflect.Value) {
	switch typ {
	case typeChar:
		b := m.Byte()
		if v.IsValid() {
			v.SetUint(uint64(b))
		}
	case typeInt:
		n := m.Int()
		if v.IsValid() {
			v.SetInt(int64(n))
		}
	case typeString:
		s := string(m.Buffer())
		if v.IsValid() {
			v.SetString(s)
		}
	case typePtr:
		p := m.Pointer()
		if v.IsValid() {
			v.SetUint(uint64(p))
		}
	case typeTime:
		t := m.Time()
		if v.IsValid() {
			v.Set(reflect.ValueOf(t))
		}
	case typeArray:
		elem := m.GetType()
		length := int(m.Int())
		var elemv reflect.Value
		if v.IsValid() {
			newv := reflect.MakeSlice(v.Type(), length, length)
			v.Set(newv)
		}
		for i := 0; i < length; i++ {
			if v.IsValid() {
				elemv = v.Index(i)
			}
			m.decodeValue(elem, elemv)
		}
	case typeMap:
		tkey, tval := m.GetType(), m.GetType()
		length := int(m.Int())
		var valk reflect.Value
		for i := 0; i < length; i++ {
			m.decodeValue(tkey, valk)
			m.decodeValue(tval, valk)
		}
	default:
		panic("unknown type " + typ)
	}

}
