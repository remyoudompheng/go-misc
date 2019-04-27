package rdvhash

import (
	"fmt"
	"testing"
)

func TestBalance(t *testing.T) {
	var keys []string
	indices := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := 0; i < 10000; i++ {
		keys = append(keys, fmt.Sprint(i))
	}

	{
		t.Logf("balancing on 10 slots")
		var stats [10]int
		for _, k := range keys {
			s := Shuffle(k, indices)
			stats[s[0]]++
		}
		t.Log(stats)
		for i, c := range stats {
			if c < 950 {
				t.Errorf("slot %d is unbalanced (%d items)", i, c)
			}
		}
	}
	{
		t.Logf("balancing on 8 slots (9 and 10 are down)")
		var stats [8]int
		for _, k := range keys {
			s := Shuffle(k, indices)
			for _, i := range s {
				if i < 8 {
					stats[i]++
					break
				}
			}
		}
		t.Log(stats)
		for i, c := range stats {
			if c < 1200 {
				t.Errorf("slot %d is unbalanced (%d items)", i, c)
			}
		}
	}

	indices = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		10, 11, 12, 13, 14, 15, 16, 17}
	{
		t.Logf("balancing on 18 slots")
		var stats [18]int
		for _, k := range keys {
			s := Shuffle(k, indices)
			stats[s[0]]++
		}
		t.Log(stats)
		for i, c := range stats {
			if c < 10000/20 {
				t.Errorf("slot %d is unbalanced (%d items)", i, c)
			}
		}
	}
	{
		t.Logf("balancing on 12 slots (6 are down)")
		var stats [12]int
		for _, k := range keys {
			s := Shuffle(k, indices)
			for _, i := range s {
				if i < 12 {
					stats[i]++
					break
				}
			}
		}
		t.Log(stats)
		for i, c := range stats {
			if c < 10000/14 {
				t.Errorf("slot %d is unbalanced (%d items)", i, c)
			}
		}
	}
}

func BenchmarkShuffle(b *testing.B) {
	const key = "supercalifragilisticexpialidocious"
	indices := []int{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
	}

	for i := 0; i < b.N; i++ {
		_ = Shuffle(key, indices)
	}
}
