package weechat

import (
	"flag"
	"reflect"
	"testing"
)

var external = flag.Bool("external", false, "use external")

func TestExternal(t *testing.T) {
	if !*external {
		t.Logf("skipping.")
		return
	}

	c, err := Dial("localhost:12001")
	if err != nil {
		t.Fatalf("dial: %s", err)
	}

	err = c.send(cmdInit)
	if err != nil {
		t.Errorf("init: %s", err)
	}

	// nicklist.
	err = c.send(cmdNicklist)
	if err != nil {
		t.Errorf("nicklist: %s", err)
	}

	s, err := c.recv()
	if err != nil {
		t.Fatalf("read: %s", err)
	}
	msg := message(s)
	id, typ := msg.Buffer(), msg.GetType()
	t.Logf("id=%s type=%v", id, typ)
	var nicks []Nick
	msg.Hdata(reflect.ValueOf(&nicks).Elem())
	if len(nicks) > 50 {
		nicks = nicks[:50]
	}
	t.Logf("%+v...", nicks)

	// get buffer list.
	err = c.send(cmdHdata, "buffer:gui_buffers(*)")
	if err == nil {
		s, err = c.recv()
	}
	if err != nil {
		t.Fatalf("buffer list: %s", err)
	}
	msg = message(s)
	id, typ = msg.Buffer(), msg.GetType()
	t.Logf("id=%s type=%v", id, typ)
	var buflist []Buffer
	msg.Hdata(reflect.ValueOf(&buflist).Elem())
	t.Logf("%+v", buflist)

	// lines.
	err = c.send(cmdHdata, "buffer:gui_buffers(*)/lines/first_line(*)/data")
	if err == nil {
		s, err = c.recv()
	}
	if err != nil {
		t.Fatalf("buffer data: %s", err)
	}
	msg = message(s)
	id, typ = msg.Buffer(), msg.GetType()
	t.Logf("id=%s type=%v", id, typ)
	var lines []LineData
	msg.Hdata(reflect.ValueOf(&lines).Elem())
	if len(lines) > 50 {
            lines = lines[:50]
	}
      for i := range lines {
            lines[i].Clean()
      }
	t.Logf("%+v", lines)
}
