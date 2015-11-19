package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestEmptyPdf(t *testing.T) {
	f, err := ioutil.TempFile("", "pdftest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	p, err := NewPDFWriter(f)
	if err != nil {
		t.Fatal(err)
	}
	p.WriteInfo("test document", time.Now())
	p.WritePage(21*CM, 29.7*CM,
		[]byte("q\n173.52 0 0 245.76 0 0 cm\nQ\n"))
	p.Flush()
	if p.err != nil {
		t.Fatal(p.err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("pdfinfo", f.Name()).CombinedOutput()
	t.Logf("%s", out)
	if err != nil {
		t.Error(err)
	}
}
