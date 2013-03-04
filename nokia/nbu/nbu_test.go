package nbu

import (
	"strings"
	"testing"
)

func TestReadTime(t *testing.T) {
	var input = "\x61\x18\xce\x01\x40\xde\x71\x9e"
	tm, err := readTime(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if tm.String() != "2013-03-03 22:51:18.948 +0000 UTC" {
		t.Errorf("got %s, expected %s", tm,
			"2013-03-03 22:51:18.948 +0000 UTC")
	}
}

func TestReadString(t *testing.T) {
	var input = "\x05\x00C\x003\x00-\x000\x000\x00"
	s, err := readString(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if s != "C3-00" {
		t.Errorf("got %q, expected %q", s, "C3-00")
	}
}
