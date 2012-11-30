package go8

import (
	"bytes"
	"fmt"
	"go/token"
)

type Prog struct {
	Op   int
	Name string
	Line int
	Pos  token.Position // interpreted line number

	From Addr
	To   Addr
}

func (p Prog) String() string {
	if p.Name != "" {
		return fmt.Sprintf("%-8s %q (:%d)", opnames[p.Op], p.Name, p.Line)
	}
	pos := "(" + p.Pos.String() + ")"
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%24s %-8s %s", pos, opnames[p.Op], p.From)
	if p.To.Type != D_NONE {
		fmt.Fprintf(buf, ",%s", p.To)
	}
	return buf.String()
}

type Addr struct {
	Type    int
	Sym     string
	Index   int
	Scale   int
	Offset  int32
	Offset2 int32

	GoType string

	FloatIEEE uint64
	StringVal [8]byte
}

func (a Addr) String() string {
	idxsuf := ""
	if a.Type != D_ADDR && a.Index != D_NONE {
		idxsuf = fmt.Sprintf("(%s*%d)", regnames[a.Index], a.Scale)
	}
	// Registers.
	if a.Type >= D_INDIR {
		if a.Offset != 0 {
			return fmt.Sprintf("%d(%s)%s", a.Offset, regnames[a.Type-D_INDIR], idxsuf)
		} else {
			return "(" + regnames[a.Type-D_INDIR] + ")" + idxsuf
		}
	}
	if D_AL <= a.Type && a.Type <= D_GS {
		if a.Offset != 0 {
			return fmt.Sprintf("%d,%s%s", a.Offset, regnames[a.Type], idxsuf)
		} else {
			return regnames[a.Type] + idxsuf
		}
	}
	// Addresses.
	switch a.Type {
	case D_NONE:
		return ""
	case D_EXTERN:
		return fmt.Sprintf("%s+%d(SB)%s", a.Sym, a.Offset, idxsuf)
	case D_AUTO:
		return fmt.Sprintf("%s+%d(SP)%s", a.Sym, a.Offset, idxsuf)
	case D_PARAM:
		return fmt.Sprintf("%s+%d(FP)%s", a.Sym, a.Offset, idxsuf)
	case D_CONST:
		// integer immediate
		return fmt.Sprintf("$%d%s", a.Offset, idxsuf)
	case D_CONST2:
		// ??
		return fmt.Sprintf("$%d-%d%s", a.Offset, a.Offset2, idxsuf)
	case D_SCONST:
		// chunk of string literal
		s := a.StringVal[:]
		s = bytes.TrimRight(s, "\x00")
		return fmt.Sprintf("$%q%s", s, idxsuf)
	case D_ADDR:
		ind := a
		ind.Type = a.Index
		ind.Index = D_NONE
		return fmt.Sprintf("$%s", ind)
	case D_BRANCH:
		return fmt.Sprintf("%d(PC)", a.Offset)
	}
	panic("ignored type " + regnames[a.Type])
}
