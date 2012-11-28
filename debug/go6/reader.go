package go6

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type Reader struct {
	rd   *bufio.Reader
	syms [256]string
}

func NewReader(r io.Reader) *Reader {
	return &Reader{rd: bufio.NewReader(r)}
}

func read2(r *bufio.Reader) (uint16, error) {
	var buf [2]byte
	_, err := r.Read(buf[:])
	return binary.LittleEndian.Uint16(buf[:]), err
}

func read4(r *bufio.Reader) (uint32, error) {
	var buf [4]byte
	_, err := r.Read(buf[:])
	return binary.LittleEndian.Uint32(buf[:]), err
}

func read8(r *bufio.Reader) (uint64, error) {
	var buf [8]byte
	_, err := r.Read(buf[:])
	return binary.LittleEndian.Uint64(buf[:]), err
}

type errOpOutOfRange int

func (e errOpOutOfRange) Error() string { return fmt.Sprintf("opcode %d out of range", int(e)) }

type errIO struct {
	When string
	Err  error
}

func (e *errIO) Error() string {
	return fmt.Sprintf("error while reading %s: %s", e.When, e.Err)
}

func (r *Reader) ReadProg() (p Prog, err error) {
	op, err := read2(r.rd)
	if err != nil {
		return
	}
	if op <= AXXX || op >= ALAST {
		return p, errOpOutOfRange(op)
	}
	switch op {
	case ANAME, ASIGNAME:
		sig := uint32(0)
		if op == ASIGNAME {
			sig, err = read4(r.rd)
			if err != nil {
				return p, &errIO{When: "SIGNAME op", Err: err}
			}
		}
		typ, err1 := r.rd.ReadByte()
		sym, err2 := r.rd.ReadByte()
		bname, err := r.rd.ReadString(0)
		switch {
		case err1 != nil:
			return p, &errIO{When: "symbol type", Err: err}
		case err2 != nil:
			return p, &errIO{When: "symbol id", Err: err}
		case err != nil:
			return p, &errIO{When: "symbol value", Err: err}
		}
		name := string(bname[:len(bname)-1])
		// Register symbol.
		r.syms[sym] = name
		_, _ = sig, typ
		return Prog{Op: int(op), Name: name}, nil
	}

	// Common instruction data.
	line, err := read4(r.rd)
	from, err1 := r.ReadAddr()
	to, err2 := r.ReadAddr()
	switch {
	case err != nil:
		return p, &errIO{When: "line number", Err: err}
	case err1 != nil:
		return p, &errIO{When: "from address", Err: err}
	case err2 != nil:
		return p, &errIO{When: "to address", Err: err}
	}

	return Prog{
		Op: int(op), Line: int(line),
		From: from, To: to}, nil
}

func (r *Reader) ReadAddr() (Addr, error) {
	a := Addr{Type: D_NONE, Index: D_NONE, Scale: 0}
	tag, err := r.rd.ReadByte()
	if tag&T_INDEX != 0 {
		var index, scale byte
		index, err = r.rd.ReadByte()
		scale, err = r.rd.ReadByte()
		a.Index, a.Scale = int(index), int(scale)
	}
	if tag&T_OFFSET != 0 {
		if tag&T_64 != 0 {
			o64, e := read8(r.rd)
			a.Offset, err = int64(o64), e
		} else {
			o32, e := read4(r.rd)
			a.Offset, err = int64(int32(o32)), e
		}
	}
	if tag&T_SYM != 0 {
		a.Sym, err = r.ReadSym()
	}
	// Constants.
	switch {
	case tag&T_FCONST != 0:
		a.Type = D_FCONST
		a.FloatIEEE, err = read8(r.rd)
	case tag&T_SCONST != 0:
		a.Type = D_SCONST
		_, err = r.rd.Read(a.StringVal[:])
	}
	// Other
	if tag&T_TYPE != 0 {
		var typ byte
		typ, err = r.rd.ReadByte()
		a.Type = int(typ)
	}
	if tag&T_GOTYPE != 0 {
		a.GoType, err = r.ReadSym()
	}
	// TODO: finish this.
	return a, err
}

func (r *Reader) ReadSym() (string, error) {
	b, err := r.rd.ReadByte()
	return r.syms[b], err
}
