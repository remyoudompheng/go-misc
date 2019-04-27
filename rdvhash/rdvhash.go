// A stupid rendezvous hashing implementation.
package rdvhash

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"
)

// Shuffle produces a random permutation of indices according
// to an input key
func Shuffle(key string, indices []int) (res []int) {
	sum := md5.Sum([]byte(key))
	h1 := binary.LittleEndian.Uint64(sum[:8])
	h2 := binary.LittleEndian.Uint64(sum[8:])
	r := rand.New(&stupidSource{h1, h2})
	res = append(res, indices...)
	r.Shuffle(len(res), func(i, j int) {
		res[i], res[j] = res[j], res[i]
	})
	return res
}

type stupidSource struct {
	x, y uint64
}

// An implementation of xorshift128+
func (s *stupidSource) Int63() int64 {
	x, y := s.x, s.y
	x ^= x << 23
	x ^= x >> 17
	x ^= y ^ (y >> 26)
	s.x, s.y = y, x
	return int64(x + y)
}

func (s *stupidSource) Seed(x int64) {
	s.x = uint64(x)
}
