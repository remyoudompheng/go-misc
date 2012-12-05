package pprof

import (
	"io"
	"os"
	"testing"
)

func TestHeapProfile(t *testing.T) {
	f, err := os.Open("testdata/heap.prof")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := NewHeapProfParser(f)
	if err != nil {
		t.Fatal(err)
	}

	syms := readSymbols("testdata/heap.prof.symbols")
	printed := 0
	for {
		rec, err := p.ReadRecord()
		if err == io.EOF {
			break
		}
		s := stringify(rec.Trace, syms)
		if printed < 10 {
			printed++
			t.Logf("%d:%d [%d:%d] @ %v",
				rec.LiveObj, rec.LiveBytes,
				rec.AllocObj, rec.AllocBytes, s)
		}
	}
}

func BenchmarkHeapProfile(b *testing.B) {
	f, err := os.Open("testdata/__prof")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	p, err := NewHeapProfParser(f)
	if err != nil {
		b.Fatal(err)
	}
	bsize := int64(0)
	for i := 0; i < b.N; i++ {
		rec, err := p.ReadRecord()
		if err == io.EOF {
			// rewind.
			bsize += tell(f)
			f.Seek(0, 0)
			p, _ = NewHeapProfParser(f)
			continue
		}
		_ = rec
	}
	bsize += tell(f)
	b.SetBytes(bsize / int64(b.N))
}
