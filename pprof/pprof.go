package main

import (
	"io"
	"log"
	"os"

	"github.com/remyoudompheng/go-misc/pprof/parser"
	"github.com/remyoudompheng/go-misc/pprof/report"
)

func LoadSymbols(exe string) (*report.Reporter, error) {
	reporter := new(report.Reporter)
	err := reporter.SetExecutable(exe)
	return reporter, err
}

func LoadProfile(r *report.Reporter, prof string) {
	f, err := os.Open(prof)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	p, err := parser.NewCpuProfParser(f)
	if err != nil {
		log.Fatal(err)
	}

	for {
		trace, count, err := p.ReadTrace()
		if trace == nil && err == io.EOF {
			break
		}
		r.Add(trace, int64(count))
	}
}

func PrintGraph(w io.Writer, r *report.Reporter, exe string) {
	g := r.GraphByFunc(report.ColCPU)
	report := report.GraphReport{
		Prog:  exe,
		Total: r.Total(report.ColCPU),
		Unit:  "samples",
		Graph: g,
	}
	report.WriteTo(w)
}

func main() {
	exe, prof := os.Args[1], os.Args[2]
	r, err := LoadSymbols(exe)
	if err != nil {
		log.Fatal(err)
	}
	LoadProfile(r, prof)
	PrintGraph(os.Stdout, r, exe)
}
