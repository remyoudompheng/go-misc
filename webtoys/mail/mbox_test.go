package mail

import (
	"os"
	"testing"
)

func TestMbox(t *testing.T) {
	f, err := os.Open("testdata/mbox")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	box, err := Open(f)
	t.Log(err)
	t.Logf("%+v", box.msgs)
}
