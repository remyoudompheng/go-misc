package pprof

import (
	"sort"
)

// A Reporter is used to produce reports from profile data.
type Reporter struct {
	Resolver  func(uint64) string
	stats     map[uint64]*Stats // stats per PC.
	freeStats []Stats           // allocation pool.
}

type Stats struct {
	Self    int64
	Cumul   int64
	Callees map[string]int64
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

func (r *Reporter) Add(trace []uint64, count int64) {
	if r.stats == nil {
		r.stats = make(map[uint64]*Stats)
	}
	// Only the last point.
	s := r.getStats(trace[0])
	s.Self += count
	// Record cumulated stats.
	seen := make(map[uint64]bool, len(trace))
	for i, a := range trace {
		s := r.getStats(a)
		if !seen[a] {
			seen[a] = true
			s.Cumul += count
		}
		if i > 0 {
			callee := trace[i-1]
			if s.Callees == nil {
				s.Callees = make(map[string]int64)
			}
			s.Callees[r.Resolver(callee)] += count
		}
	}
}

func (r *Reporter) ReportByFunc() []ReportLine {
	lines := make(map[string][2]int64, len(r.stats))
	for a, v := range r.stats {
		name := r.Resolver(a)
		s := lines[name]
		s[0] += v.Self
		s[1] += v.Cumul
		lines[name] = s
	}
	entries := make([]ReportLine, 0, len(lines))
	for name, values := range lines {
		entries = append(entries, ReportLine{Name: name,
			Self: float64(values[0]), Cumul: float64(values[1])})
	}
	sort.Sort(bySelf(entries))
	return entries
}

func (r *Reporter) ReportByPC() []ReportLine {
	return nil
}

type ReportLine struct {
	Name        string
	Self, Cumul float64
}

type bySelf []ReportLine

func (s bySelf) Len() int      { return len(s) }
func (s bySelf) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s bySelf) Less(i, j int) bool {
	return s[i].Self > s[j].Self || (s[i].Self == s[j].Self && s[i].Name < s[j].Name)
}

type byCumul []ReportLine

func (s byCumul) Len() int      { return len(s) }
func (s byCumul) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byCumul) Less(i, j int) bool {
	return s[i].Cumul > s[j].Cumul || (s[i].Self == s[j].Self && s[i].Name < s[j].Name)
}
