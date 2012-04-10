package extramath

import (
	"math/rand"
	"testing"
)

func TestDivmod(t *testing.T) {
	const N = 1e6
	for i := 0; i < N; i++ {
		a := uint64(rand.Uint32())<<32 | uint64(rand.Uint32())
		b := uint64(rand.Uint32())<<32 | uint64(rand.Uint32())
		if b == 0 {
			continue
		}
		quo, rem := Divmod(a, b)
		if quo != a/b || rem != a%b {
			t.Errorf("expected (%d, %d) got (%d, %d)", a/b, a%b, quo, rem)
		}
	}
}

func BenchmarkDivmod(bm *testing.B) {
	a, b := uint64(123456789123456789), uint64(987654321)
	for i := 0; i < bm.N; i++ {
		quo, rem := Divmod(a, b)
		_, _ = quo, rem
	}
}

func BenchmarkDivmodNaive(bm *testing.B) {
	a, b := uint64(123456789123456789), uint64(987654321)
	for i := 0; i < bm.N; i++ {
		quo, rem := a/b, a%b
		_, _ = quo, rem
	}
}

func naiveMul(a, b uint64) (hi, lo uint64) {
	ahi, alo := a>>32, uint64(uint32(a))
	bhi, blo := b>>32, uint64(uint32(b))
	carry := uint64(uint32(ahi*blo)) + uint64(uint32(bhi*alo)) + (alo*blo)>>32
	hi = ahi*bhi + (ahi*blo)>>32 + (bhi*alo)>>32 + carry>>32
	return hi, a * b
}

func TestMul(t *testing.T) {
	const N = 1e5
	for i := 0; i < N; i++ {
		a := uint64(rand.Uint32())<<32 | uint64(rand.Uint32())
		b := uint64(rand.Uint32())<<32 | uint64(rand.Uint32())
		hi, lo := Mul(a, b)
		if hi2, lo2 := naiveMul(a, b); hi2 != hi || lo2 != lo {
			t.Errorf("expected %016x%016x got %016x%016x", hi2, lo2, hi, lo)
		}
	}
}

func BenchmarkMul(bm *testing.B) {
	a, b := uint64(123456789123456789), uint64(987654321)
	for i := 0; i < bm.N; i++ {
		hi, lo := Mul(a, b)
		_, _ = hi, lo
	}
}

func BenchmarkMulNaive(bm *testing.B) {
	a, b := uint64(123456789123456789), uint64(987654321)
	for i := 0; i < bm.N; i++ {
		hi, lo := naiveMul(a, b)
		_, _ = hi, lo
	}
}
