package report

import (
	"fmt"
	"sort"
)

// A Reporter is used to produce reports from profile data.
type Reporter struct {
	Resolver  func(uint64) string
	stats     map[uint64]*Stats // stats per PC.
	freeStats []Stats           // allocation pool.
}

type Stats struct {
	Self    [4]int64
	Cumul   [4]int64
	Callees map[string][4]int64
}

func (r *Reporter) getStats(key uint64) *Stats {
	if p := r.stats[key]; p != nil {
		return p
	}
	if len(r.freeStats) == 0 {
		r.freeStats = make([]Stats, 64)
	}
	s := &r.freeStats[0]
	r.freeStats = r.freeStats[1:]
	r.stats[key] = s
	return s
}

// Add registers data for a given stack trace. There may be at most
// 4 count arguments, as needed in heap profiles.
func (r *Reporter) Add(trace []uint64, count ...int64) {
	if r.stats == nil {
		r.stats = make(map[uint64]*Stats)
	}
	if len(count) > 4 {
		err := fmt.Errorf("too many counts (%d) to register in reporter", len(count))
		panic(err)
	}
	// Only the last point.
	s := r.getStats(trace[0])
	for i, n := range count {
		s.Self[i] += n
	}
	// Record cumulated stats.
	seen := make(map[uint64]bool, len(trace))
	for i, a := range trace {
		s := r.getStats(a)
		if !seen[a] {
			seen[a] = true
			for j, n := range count {
				s.Cumul[j] += n
			}
		}
		if i > 0 {
			callee := trace[i-1]
			if s.Callees == nil {
				s.Callees = make(map[string][4]int64)
			}
			edges := s.Callees[r.Resolver(callee)]
			for j, n := range count {
				edges[j] += n
			}
			s.Callees[r.Resolver(callee)] = edges
		}
	}
}

const (
	ColCPU        = 0
	ColLiveObj    = 0
	ColLiveBytes  = 1
	ColAllocObj   = 2
	ColAllocBytes = 3
)

func (r *Reporter) ReportByFunc(column int) []ReportLine {
	lines := make(map[string][2][4]float64, len(r.stats))
	for a, v := range r.stats {
		name := r.Resolver(a)
		s := lines[name]
		for i := 0; i < 4; i++ {
			s[0][i] += float64(v.Self[i])
			s[1][i] += float64(v.Cumul[i])
		}
		lines[name] = s
	}
	entries := make([]ReportLine, 0, len(lines))
	for name, values := range lines {
		entries = append(entries, ReportLine{Name: name,
			Self: values[0], Cumul: values[1]})
	}
	sort.Sort(bySelf{entries, column})
	return entries
}

func (r *Reporter) ReportByPC() []ReportLine {
	return nil
}

type ReportLine struct {
	Name        string
	Self, Cumul [4]float64
}

type bySelf struct {
	slice []ReportLine
	col   int
}

func (s bySelf) Len() int      { return len(s.slice) }
func (s bySelf) Swap(i, j int) { s.slice[i], s.slice[j] = s.slice[j], s.slice[i] }

func (s bySelf) Less(i, j int) bool {
	left, right := s.slice[i].Self[s.col], s.slice[j].Self[s.col]
	if left > right {
		return true
	}
	if left == right {
		return s.slice[i].Name < s.slice[j].Name
	}
	return false
}

type byCumul bySelf

func (s byCumul) Len() int      { return len(s.slice) }
func (s byCumul) Swap(i, j int) { s.slice[i], s.slice[j] = s.slice[j], s.slice[i] }

func (s byCumul) Less(i, j int) bool {
	left, right := s.slice[i].Cumul[s.col], s.slice[j].Cumul[s.col]
	if left > right {
		return true
	}
	if left == right {
		return s.slice[i].Name < s.slice[j].Name
	}
	return false
}
